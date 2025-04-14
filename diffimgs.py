#!/usr/bin/python
# -*- coding: utf-8 -*-
# 比较两个图片的差异
# 任意图片先转换为灰度图
# 调整图片大小，让图片主体看起来一样大小
# 调整图片物体移动中间位置
# 计算两张灰度图的差异
# 显示两个图片的差异 并保存差异图片

import cv2
import numpy as np
import os
import argparse
from skimage.metrics import structural_similarity as ssim

def move_object_to_center(img):
    """
    改进版的物体居中函数，可处理抽象图像和渐变图像
    """
    # 复制原图以避免修改原图
    result = img.copy()
    
    # 转换为灰度图
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY) if len(img.shape) == 3 else img.copy()
    
    # 使用自适应阈值处理，更适合处理渐变和抽象图像
    thresh = cv2.adaptiveThreshold(
        gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
        cv2.THRESH_BINARY, 11, 2
    )
    
    # 获取图像质心作为参考点
    moments = cv2.moments(thresh)
    
    # 如果无法计算质心，则使用图像中心
    if moments["m00"] == 0:
        cX = img.shape[1] // 2
        cY = img.shape[0] // 2
    else:
        cX = int(moments["m10"] / moments["m00"])
        cY = int(moments["m01"] / moments["m00"])
    
    # 计算需要移动的距离
    img_center_x = img.shape[1] // 2
    img_center_y = img.shape[0] // 2
    
    dx = img_center_x - cX
    dy = img_center_y - cY
    
    # 创建平移矩阵
    M = np.float32([[1, 0, dx], [0, 1, dy]])
    
    # 应用平移变换
    shifted = cv2.warpAffine(result, M, (img.shape[1], img.shape[0]))
    
    return shifted

def calculate_similarity(diff):
    """
    计算两图片的相似度
    返回0-100的相似度值，100表示完全相同
    """
    # 计算非零像素的数量（表示差异）
    non_zero_count = np.count_nonzero(diff)
    # 计算总像素数
    total_pixels = diff.shape[0] * diff.shape[1]
    
    # 相似度 = (1 - 差异像素占比) * 100
    if total_pixels == 0:
        return 100.0
    
    similarity = (1 - (non_zero_count / total_pixels)) * 100
    return similarity

def calculate_ssim_similarity(img1, img2):
    """
    使用SSIM计算结构相似度
    返回0-100的相似度值，100表示完全相同
    """
    # 确保两张图片是灰度图
    if len(img1.shape) == 3:
        img1 = cv2.cvtColor(img1, cv2.COLOR_BGR2GRAY)
    if len(img2.shape) == 3:
        img2 = cv2.cvtColor(img2, cv2.COLOR_BGR2GRAY)
    
    # 计算SSIM
    score, _ = ssim(img1, img2, full=True)
    
    # 转换为0-100的范围
    return score * 100

def create_difference_image(diff, threshold=30):
    """
    创建黑色背景的差异图，差异部分用红色标出
    
    参数:
        diff: 差异图像
        threshold: 差异阈值，大于该值视为差异点
    
    返回:
        标记了差异的彩色图像
    """
    # 创建黑色背景
    height, width = diff.shape
    diff_image = np.zeros((height, width, 3), dtype=np.uint8)
    
    # 设置差异部分为红色 (B,G,R) = (0,0,255)
    # 仅当差异超过阈值时才标记为差异
    diff_mask = diff > threshold
    diff_image[diff_mask, 2] = 255  # 红色通道设为255
    
    return diff_image

def preprocess_cad_image(img, keep_original_clarity=True):
    """
    对CAD图像进行预处理，可选择保持原始清晰度或增强线条
    
    参数:
        img: 输入图像
        keep_original_clarity: 是否保持原始清晰度，True只进行基本黑白转换，False进行线条增强
        
    返回:
        预处理后的图像
    """
    # 复制原图以避免修改原图
    result = img.copy()
    
    # 转换为灰度图
    if len(img.shape) == 3:
        gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    else:
        gray = img.copy()
    
    if keep_original_clarity:
        # 简单的二值化处理，保持线条清晰
        # 使用OTSU自动确定阈值，适应不同亮度的图像
        _, binary = cv2.threshold(gray, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
    else:
        # 增强模式 - 应用自适应阈值处理，增强线条
        binary = cv2.adaptiveThreshold(
            gray, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, 
            cv2.THRESH_BINARY, 9, 2
        )
        
        # 使用形态学操作增强线条
        kernel = np.ones((2, 2), np.uint8)
        dilated = cv2.dilate(binary, kernel, iterations=1)
        
        # 应用边缘检测增强线条
        edges = cv2.Canny(dilated, 50, 150)
        
        # 合并边缘和二值图像
        binary = cv2.bitwise_or(binary, edges)
    
    # 转回三通道图像以便后续处理
    if len(img.shape) == 3:
        processed = cv2.cvtColor(binary, cv2.COLOR_GRAY2BGR)
        return processed
    else:
        return binary

# 添加一个新的对齐函数
def align_images(img1, img2):
    """
    使用多种方法尝试对齐两张图片，支持平移、旋转和缩放变换的不变性处理
    """
    # 保存原始图像用于回退
    orig_img1 = img1.copy()
    orig_img2 = img2.copy()
    
    # 转换为灰度图，以便特征检测
    gray1 = cv2.cvtColor(img1, cv2.COLOR_BGR2GRAY) if len(img1.shape) == 3 else img1.copy()
    gray2 = cv2.cvtColor(img2, cv2.COLOR_BGR2GRAY) if len(img2.shape) == 3 else img2.copy()
    
    # 1. 使用ORB特征检测器进行特征点匹配（最适合处理几何变换）
    try:
        # 创建ORB检测器，增加特征点数量以提高匹配成功率
        orb = cv2.ORB_create(nfeatures=2000, scaleFactor=1.2, nlevels=8)
        
        # 检测关键点和计算描述符
        kp1, des1 = orb.detectAndCompute(gray1, None)
        kp2, des2 = orb.detectAndCompute(gray2, None)
        
        # 如果找到了足够的特征点
        if des1 is not None and des2 is not None and len(des1) >= 4 and len(des2) >= 4:
            # 创建特征匹配器
            bf = cv2.BFMatcher(cv2.NORM_HAMMING, crossCheck=True)
            
            # 匹配特征点
            matches = bf.match(des1, des2)
            
            # 按距离排序
            matches = sorted(matches, key=lambda x: x.distance)
            
            # 保留最佳匹配
            good_matches = matches[:min(100, len(matches))]
            
            if len(good_matches) >= 4:  # 需要至少4个点来计算单应矩阵
                # 提取匹配点的坐标
                src_pts = np.float32([kp1[m.queryIdx].pt for m in good_matches]).reshape(-1, 1, 2)
                dst_pts = np.float32([kp2[m.trainIdx].pt for m in good_matches]).reshape(-1, 1, 2)
                
                # 计算单应性矩阵（支持平移、旋转和缩放变换）
                H, mask = cv2.findHomography(src_pts, dst_pts, cv2.RANSAC, 5.0)
                
                if H is not None:
                    # 应用变换，将img1变换到img2的视角
                    h, w = img2.shape[:2]
                    aligned_img1 = cv2.warpPerspective(img1, H, (w, h))
                    
                    # 计算变换后的相似度，确认对齐效果
                    gray_aligned = cv2.cvtColor(aligned_img1, cv2.COLOR_BGR2GRAY) if len(aligned_img1.shape) == 3 else aligned_img1
                    ssim_score = calculate_ssim_similarity(gray_aligned, gray2) / 100.0
                    
                    if ssim_score > 0.3:  # 如果相似度足够高，使用这个对齐结果
                        return aligned_img1, img2
    except Exception as e:
        print(f"ORB特征匹配失败: {e}")
    
    # 2. 如果ORB特征匹配失败，尝试使用ECC增强型相关系数方法（适用于平移、旋转和仿射变换）
    try:
        # 定义要使用的变换类型
        warp_mode = cv2.MOTION_AFFINE  # 支持平移、旋转、缩放和剪切
        
        # 指定ECC算法的迭代次数和终止条件
        criteria = (cv2.TERM_CRITERIA_EPS | cv2.TERM_CRITERIA_COUNT, 2000, 1e-8)
        
        # 创建变换矩阵
        if warp_mode == cv2.MOTION_HOMOGRAPHY:
            warp_matrix = np.eye(3, 3, dtype=np.float32)
        else:
            warp_matrix = np.eye(2, 3, dtype=np.float32)
        
        # 运行ECC算法
        _, warp_matrix = cv2.findTransformECC(
            gray2, gray1, warp_matrix, warp_mode, criteria, None, 5
        )
        
        # 应用变换
        h, w = img2.shape[:2]
        if warp_mode == cv2.MOTION_HOMOGRAPHY:
            aligned_img1 = cv2.warpPerspective(
                img1, warp_matrix, (w, h),
                flags=cv2.INTER_LINEAR + cv2.WARP_INVERSE_MAP,
                borderMode=cv2.BORDER_CONSTANT
            )
        else:
            aligned_img1 = cv2.warpAffine(
                img1, warp_matrix, (w, h),
                flags=cv2.INTER_LINEAR + cv2.WARP_INVERSE_MAP,
                borderMode=cv2.BORDER_CONSTANT
            )
        
        # 计算变换后的相似度，确认对齐效果
        gray_aligned = cv2.cvtColor(aligned_img1, cv2.COLOR_BGR2GRAY) if len(aligned_img1.shape) == 3 else aligned_img1
        ssim_score = calculate_ssim_similarity(gray_aligned, gray2) / 100.0
        
        if ssim_score > 0.3:  # 如果相似度足够高，使用这个对齐结果
            return aligned_img1, img2
    except Exception as e:
        print(f"ECC对齐失败: {e}")
    
    # 3. 如果之前方法都失败，尝试使用Hu矩匹配（适用于形状匹配，具有旋转和缩放不变性）
    try:
        # 提取轮廓
        _, thresh1 = cv2.threshold(gray1, 127, 255, cv2.THRESH_BINARY)
        _, thresh2 = cv2.threshold(gray2, 127, 255, cv2.THRESH_BINARY)
        
        contours1, _ = cv2.findContours(thresh1, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        contours2, _ = cv2.findContours(thresh2, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        if contours1 and contours2:
            # 找到面积最大的轮廓
            c1 = max(contours1, key=cv2.contourArea)
            c2 = max(contours2, key=cv2.contourArea)
            
            # 计算Hu矩（具有旋转、平移和缩放不变性）
            hu1 = cv2.HuMoments(cv2.moments(c1)).flatten()
            hu2 = cv2.HuMoments(cv2.moments(c2)).flatten()
            
            # 计算矩的距离
            dist = np.sum(np.abs(np.log(hu1 + 1e-10) - np.log(hu2 + 1e-10)))
            
            if dist < 1.0:  # 如果矩的距离较小，说明形状相似
                # 获取轮廓的最小外接矩形
                rect1 = cv2.minAreaRect(c1)
                rect2 = cv2.minAreaRect(c2)
                
                # 获取矩形的中心、大小和角度
                (cx1, cy1), (w1, h1), angle1 = rect1
                (cx2, cy2), (w2, h2), angle2 = rect2
                
                # 计算缩放比例
                scale_x = w2 / w1 if w1 > 0 else 1.0
                scale_y = h2 / h1 if h1 > 0 else 1.0
                scale = (scale_x + scale_y) / 2.0
                
                # 计算旋转角度差异
                angle_diff = angle2 - angle1
                
                # 计算平移
                dx = cx2 - cx1
                dy = cy2 - cy1
                
                # 创建复合变换矩阵
                # 先平移到原点，然后缩放和旋转，最后平移到目标位置
                M = cv2.getRotationMatrix2D((cx1, cy1), angle_diff, scale)
                M[0, 2] += dx
                M[1, 2] += dy
                
                # 应用变换
                h, w = img2.shape[:2]
                aligned_img1 = cv2.warpAffine(img1, M, (w, h))
                
                # 计算变换后的相似度
                gray_aligned = cv2.cvtColor(aligned_img1, cv2.COLOR_BGR2GRAY) if len(aligned_img1.shape) == 3 else aligned_img1
                ssim_score = calculate_ssim_similarity(gray_aligned, gray2) / 100.0
                
                if ssim_score > 0.3:  # 如果相似度足够高，使用这个对齐结果
                    return aligned_img1, img2
    except Exception as e:
        print(f"Hu矩匹配失败: {e}")
    
    # 4. 都失败的情况下，使用简单的中心对齐作为回退方案
    centered_img1 = move_object_to_center(orig_img1)
    centered_img2 = move_object_to_center(orig_img2)
    
    return centered_img1, centered_img2

def extract_and_compare_contours(img1, img2, threshold=0.8):
    """
    提取并比较两张图片的轮廓特征
    
    参数:
        img1: 第一张图片
        img2: 第二张图片
        threshold: 轮廓匹配阈值，默认0.8
        
    返回:
        similarity: 轮廓相似度(0-100)
        matched_contours: 匹配的轮廓对列表
    """
    # 转换为灰度图
    gray1 = cv2.cvtColor(img1, cv2.COLOR_BGR2GRAY) if len(img1.shape) == 3 else img1.copy()
    gray2 = cv2.cvtColor(img2, cv2.COLOR_BGR2GRAY) if len(img2.shape) == 3 else img2.copy()
    
    # 二值化处理
    _, thresh1 = cv2.threshold(gray1, 127, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
    _, thresh2 = cv2.threshold(gray2, 127, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
    
    # 提取轮廓
    contours1, _ = cv2.findContours(thresh1, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    contours2, _ = cv2.findContours(thresh2, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    
    # 过滤太小的轮廓
    min_contour_area = 100
    contours1 = [c for c in contours1 if cv2.contourArea(c) > min_contour_area]
    contours2 = [c for c in contours2 if cv2.contourArea(c) > min_contour_area]
    
    if not contours1 or not contours2:
        return 0, []
    
    # 计算每个轮廓的Hu矩
    moments1 = [cv2.HuMoments(cv2.moments(c)).flatten() for c in contours1]
    moments2 = [cv2.HuMoments(cv2.moments(c)).flatten() for c in contours2]
    
    # 对每个轮廓对计算相似度
    matched_contours = []
    used_indices = set()
    total_similarity = 0
    
    for i, (c1, m1) in enumerate(zip(contours1, moments1)):
        best_match = None
        best_similarity = -float('inf')
        best_idx = -1
        
        for j, (c2, m2) in enumerate(zip(contours2, moments2)):
            if j in used_indices:
                continue
                
            # 计算Hu矩的距离
            hu_distance = -np.sum([np.abs(np.log(abs(h1) + 1e-10) - np.log(abs(h2) + 1e-10)) 
                                 for h1, h2 in zip(m1, m2)])
            
            # 计算轮廓面积比
            area1 = cv2.contourArea(c1)
            area2 = cv2.contourArea(c2)
            area_ratio = min(area1, area2) / max(area1, area2)
            
            # 计算轮廓周长比
            peri1 = cv2.arcLength(c1, True)
            peri2 = cv2.arcLength(c2, True)
            peri_ratio = min(peri1, peri2) / max(peri1, peri2)
            
            # 综合相似度分数
            similarity = (hu_distance + area_ratio + peri_ratio) / 3
            
            if similarity > best_similarity:
                best_similarity = similarity
                best_match = c2
                best_idx = j
        
        if best_similarity > threshold:
            matched_contours.append((c1, best_match))
            used_indices.add(best_idx)
            total_similarity += best_similarity
    
    # 计算整体相似度得分(0-100)
    if matched_contours:
        avg_similarity = (total_similarity / len(matched_contours)) * 100
    else:
        avg_similarity = 0
    
    return avg_similarity, matched_contours

def visualize_contour_matches(img1, img2, matched_contours, output_path):
    """
    可视化轮廓匹配结果
    """
    h1, w1 = img1.shape[:2]
    h2, w2 = img2.shape[:2]
    
    # 创建展示图
    vis = np.zeros((max(h1, h2), w1 + w2, 3), dtype=np.uint8)
    vis[:h1, :w1] = img1
    vis[:h2, w1:w1+w2] = img2
    
    # 为每对匹配的轮廓随机生成颜色
    for c1, c2 in matched_contours:
        color = tuple(np.random.randint(0, 255, 3).tolist())
        
        # 绘制第一张图的轮廓
        cv2.drawContours(vis[:, :w1], [c1], -1, color, 2)
        
        # 绘制第二张图的轮廓（需要调整x坐标）
        c2_shifted = c2.copy()
        c2_shifted[:, :, 0] += w1
        cv2.drawContours(vis, [c2_shifted], -1, color, 2)
        
        # 连接匹配的轮廓中心点
        m1 = cv2.moments(c1)
        m2 = cv2.moments(c2)
        if m1['m00'] != 0 and m2['m00'] != 0:
            cx1 = int(m1['m10'] / m1['m00'])
            cy1 = int(m1['m01'] / m1['m00'])
            cx2 = int(m2['m10'] / m2['m00']) + w1
            cy2 = int(m2['m01'] / m2['m00'])
            cv2.line(vis, (cx1, cy1), (cx2, cy2), color, 1, cv2.LINE_AA)
    
    cv2.imwrite(output_path, vis)
    return vis

def compare_images(img1_path, img2_path, output_dir="output", threshold=90, show_result=True, 
                  align_method="auto", diff_threshold=30, is_cad=False, cad_enhance_mode=False,
                  contour_mode=False, contour_threshold=0.8):
    """
    比较两张图片并判断是否相似，支持几何不变性比较
    
    新增参数:
        contour_mode: 是否使用轮廓比对模式
        contour_threshold: 轮廓匹配阈值(0-1)，默认0.8
    """
    # 检查图片路径是否存在
    if not os.path.exists(img1_path) or not os.path.exists(img2_path):
        raise FileNotFoundError("图片路径不存在")
    
    # 创建输出目录
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    # 读取图片
    img1 = cv2.imread(img1_path)
    img2 = cv2.imread(img2_path)
    
    if img1 is None or img2 is None:
        raise ValueError("无法读取图片")
    
    # 备份原始图片用于显示
    orig_img1 = img1.copy()
    orig_img2 = img2.copy()
    
    # 预处理图像
    if is_cad:
        if cad_enhance_mode:
            print("使用CAD图像增强预处理模式...")
        else:
            print("使用CAD图像简单预处理模式，保持原图清晰度...")
            
        img1 = preprocess_cad_image(img1, keep_original_clarity=not cad_enhance_mode)
        img2 = preprocess_cad_image(img2, keep_original_clarity=not cad_enhance_mode)
        
        # 为CAD图像自动降低差异阈值
        if diff_threshold > 15:
            diff_threshold = 15
            print(f"为CAD图像自动调整差异阈值为: {diff_threshold}")
            
        # 强制使用center对齐方法，因为CAD图像特征点可能不足
        if align_method != "center":
            align_method = "center"
            print("为CAD图像自动切换到center对齐方法")
    
    # 确保两张图片大小相同
    h1, w1 = img1.shape[:2]
    h2, w2 = img2.shape[:2]
    
    # 使用较大的尺寸，以避免图像细节丢失
    target_size = (max(w1, w2), max(h1, h2))
    
    img1 = cv2.resize(img1, target_size)
    img2 = cv2.resize(img2, target_size)
    
    # 根据选择的方法对齐图像
    if align_method == "center":
        aligned_img1 = move_object_to_center(img1)
        img2 = move_object_to_center(img2)
    elif align_method == "feature":
        aligned_img1, img2 = align_images(img1, img2)
    else:  # auto
        # 先尝试使用特征匹配方法对齐图像
        aligned_img1, img2 = align_images(img1, img2)
    
    # 保存对齐后的图片以便检查
    cv2.imwrite(os.path.join(output_dir, f"{os.path.basename(img1_path)}_aligned.png"), aligned_img1)
    cv2.imwrite(os.path.join(output_dir, f"{os.path.basename(img2_path)}_aligned.png"), img2)
    
    # 转换为灰度图
    gray1 = cv2.cvtColor(aligned_img1, cv2.COLOR_BGR2GRAY)
    gray2 = cv2.cvtColor(img2, cv2.COLOR_BGR2GRAY)
    
    # 计算结构相似度(SSIM)
    ssim_similarity = calculate_ssim_similarity(gray1, gray2)
    
    # 计算两张灰度图的绝对差异
    diff = cv2.absdiff(gray1, gray2)
    
    # 计算像素相似度分数
    pixel_similarity = calculate_similarity(diff)
    
    # 如果启用轮廓比对模式
    if contour_mode:
        print("使用轮廓比对模式...")
        contour_similarity, matched_contours = extract_and_compare_contours(
            aligned_img1, img2, contour_threshold
        )
        
        # 生成轮廓匹配可视化结果
        if matched_contours:
            vis_path = os.path.join(output_dir, f"{os.path.splitext(os.path.basename(img1_path))[0]}_{os.path.splitext(os.path.basename(img2_path))[0]}_contour_matches.png")
            visualize_contour_matches(aligned_img1, img2, matched_contours, vis_path)
            print(f"轮廓匹配结果已保存至: {vis_path}")
        
        # 将轮廓相似度纳入最终相似度计算
        if is_cad:
            # 对于CAD图纸，给予轮廓相似度更高的权重
            similarity = 0.5 * contour_similarity + 0.3 * ssim_similarity + 0.2 * pixel_similarity
        else:
            # 对于普通图片，保持原有权重分配
            similarity = 0.2 * contour_similarity + 0.5 * ssim_similarity + 0.3 * pixel_similarity
        
        # 更新结果文件
        with open(os.path.join(output_dir, f"{os.path.splitext(os.path.basename(img1_path))[0]}_{os.path.splitext(os.path.basename(img2_path))[0]}_result.txt"), 'a', encoding='utf-8') as f:
            f.write(f"\n轮廓相似度: {contour_similarity:.2f}%\n")
            f.write(f"匹配轮廓数量: {len(matched_contours)}\n")
        
        print(f"轮廓相似度: {contour_similarity:.2f}%")
        print(f"匹配轮廓数量: {len(matched_contours)}")
    
    # 组合两种相似度指标，给予SSIM更高权重(70%)
    similarity = 0.7 * ssim_similarity + 0.3 * pixel_similarity
    is_similar = similarity >= threshold
    
    # 保存结果
    img1_name = os.path.basename(img1_path)
    img2_name = os.path.basename(img2_path)
    basename = f"{os.path.splitext(img1_name)[0]}_{os.path.splitext(img2_name)[0]}"
    
    # 创建并保存自定义差异图（黑色背景，红色差异）
    diff_image = create_difference_image(diff, diff_threshold)
    cv2.imwrite(os.path.join(output_dir, f"{basename}_diff_red.png"), diff_image)
    
    # 同时保存热力图版本的差异图，便于不同分析需求
    diff_colored = cv2.applyColorMap(diff, cv2.COLORMAP_JET)
    cv2.imwrite(os.path.join(output_dir, f"{basename}_diff_heatmap.png"), diff_colored)
    
    # 创建并保存对比图
    h, w = gray1.shape
    comparison = np.zeros((h, w*3, 3), dtype=np.uint8)
    
    # 展示对齐后的图像
    comparison[:, :w] = cv2.cvtColor(gray1, cv2.COLOR_GRAY2BGR)
    comparison[:, w:2*w] = cv2.cvtColor(gray2, cv2.COLOR_GRAY2BGR)
    comparison[:, 2*w:] = diff_image  # 使用红色差异图
    
    # 添加相似度信息到图片上（仅使用数字和英文，避免中文乱码）
    if is_similar:
        status_text = "SIMILAR"
    else:
        status_text = "DIFFERENT"
    
    cv2.putText(comparison, f"Similarity: {similarity:.2f}% ({status_text})", 
                (10, 30), cv2.FONT_HERSHEY_SIMPLEX, 0.8, (0, 255, 0), 2)
    
    # 添加SSIM分数
    cv2.putText(comparison, f"SSIM: {ssim_similarity:.2f}%", 
                (10, 60), cv2.FONT_HERSHEY_SIMPLEX, 0.8, (0, 255, 0), 2)
    
    cv2.imwrite(os.path.join(output_dir, f"{basename}_comparison.png"), comparison)
    
    # 保存相似度数据到文本文件（使用UTF-8编码确保中文正常显示）
    with open(os.path.join(output_dir, f"{basename}_result.txt"), 'w', encoding='utf-8') as f:
        f.write(f"图片1: {img1_path}\n")
        f.write(f"图片2: {img2_path}\n")
        f.write(f"结构相似度(SSIM): {ssim_similarity:.2f}%\n")
        f.write(f"像素相似度: {pixel_similarity:.2f}%\n")
        f.write(f"综合相似度: {similarity:.2f}%\n")
        f.write(f"判断结果: {'相似' if is_similar else '不相似'} (阈值: {threshold}%)\n")
    
    # 显示结果
    if show_result:
        cv2.imshow("Comparison Result", comparison)
        cv2.waitKey(0)
        cv2.destroyAllWindows()
    
    # 打印中文结果到控制台
    print(f"结构相似度(SSIM): {ssim_similarity:.2f}%")
    print(f"像素相似度: {pixel_similarity:.2f}%")
    print(f"综合相似度: {similarity:.2f}%")
    print(f"判断结果: {'相似' if is_similar else '不相似'} (阈值: {threshold}%)")
    
    return similarity, is_similar

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='比较两张图片的相似度，支持几何不变性对比')
    parser.add_argument('img1', help='第一张图片路径')
    parser.add_argument('img2', help='第二张图片路径')
    parser.add_argument('--output', '-o', default='output', help='输出目录路径')
    parser.add_argument('--threshold', '-t', type=float, default=90.0, help='相似度阈值(0-100)，默认90')
    parser.add_argument('--align', '-a', choices=['auto', 'center', 'feature'], default='auto', 
                        help='对齐方法: auto(自动), center(居中), feature(特征匹配)')
    parser.add_argument('--diff-threshold', '-dt', type=int, default=30, 
                        help='差异阈值，大于该值的像素被标记为差异，默认30')
    parser.add_argument('--cad', '-c', action='store_true', 
                        help='启用CAD图像处理模式（自动优化参数和预处理步骤）')
    parser.add_argument('--cad-enhance', '-ce', action='store_true',
                        help='启用CAD线条增强模式，否则只转黑白保持原图清晰度')
    parser.add_argument('--contour', '-cnt', action='store_true',
                        help='启用轮廓比对模式')
    parser.add_argument('--contour-threshold', '-ct', type=float, default=0.8,
                        help='轮廓匹配阈值(0-1)，默认0.8')
    
    args = parser.parse_args()
    
    try:
        compare_images(
            args.img1, 
            args.img2, 
            output_dir=args.output,
            threshold=args.threshold,
            show_result=False,
            align_method=args.align,
            diff_threshold=args.diff_threshold,
            is_cad=args.cad,
            cad_enhance_mode=args.cad_enhance,
            contour_mode=args.contour,
            contour_threshold=args.contour_threshold
        )
        
    except Exception as e:
        print(f"出错: {str(e)}")
