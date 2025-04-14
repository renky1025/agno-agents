from PIL import Image, ImageDraw, ImageFont
import easyocr
from translator import TranslatorService
import numpy as np
import cv2
import os
from typing import Tuple, Optional
import logging
# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class ImageTranslator:
    def __init__(self):
        self.translator = TranslatorService()
        self.supported_languages = {
            'en': 'english',
            'zh': 'chinese',
            'ja': 'japanese',
            'ko': 'korean',
            'ru': 'russian',
            'fr': 'french',
            'de': 'german',
            'es': 'spanish'
        }
        
        # 为不同语言设置默认字体
        self.fonts = {
            'zh': "C:\\Windows\\Fonts\\msyh.ttc",  # 微软雅黑
            'ja': "C:\\Windows\\Fonts\\msgothic.ttc",  # MS Gothic
            'ko': "C:\\Windows\\Fonts\\malgun.ttf",  # Malgun Gothic
            'default': "C:\\Windows\\Fonts\\Arial.ttf"  # 默认字体
        }

    def get_font_path(self, lang: str) -> str:
        """获取适合目标语言的字体路径"""
        font_path = self.fonts.get(lang, self.fonts['default'])
        if not os.path.exists(font_path):
            logger.warning(f"找不到首选字体 {font_path}，使用系统默认字体")
            font_path = self.fonts['default']
        return font_path

    def calculate_font_size(self, text: str, max_width: float, max_height: float, 
                          font_path: str, start_size: int = 12) -> int:
        """自动计算合适的字体大小"""
        font_size = start_size
        while True:
            font = ImageFont.truetype(font_path, font_size)
            text_bbox = font.getbbox(text)
            if text_bbox is None:
                return start_size
            
            text_width = text_bbox[2] - text_bbox[0]
            text_height = text_bbox[3] - text_bbox[1]
            
            if text_width > max_width or text_height > max_height:
                return max(font_size - 1, 8)  # 不要小于8号字
            
            if font_size >= 36:  # 设置最大字号
                return 36
                
            font_size += 1

    def get_background_color(self, image: Image.Image, bbox: list) -> Tuple[int, int, int]:
        """智能获取背景颜色"""
        x1, y1 = map(int, bbox[0])
        x2, y2 = map(int, bbox[2])
        
        # 转换为numpy数组以便处理
        region = np.array(image.crop((x1, y1, x2, y2)))
        
        # 使用K-means获取主要颜色
        pixels = region.reshape(-1, 3)
        pixels = np.float32(pixels)
        
        criteria = (cv2.TERM_CRITERIA_EPS + cv2.TERM_CRITERIA_MAX_ITER, 200, .1)
        flags = cv2.KMEANS_RANDOM_CENTERS
        _, labels, palette = cv2.kmeans(pixels, 2, None, criteria, 10, flags)
        
        # 返回出现次数最多的颜色
        dominant_color = np.uint8(palette[np.argmax(np.bincount(labels.flatten()))])
        return tuple(dominant_color)

    def translate_text(self, text: str, source_lang: str, target_lang: str) -> str:
        """翻译文本并处理错误"""
        try:
            translated = self.translator.translate(text, source_lang, target_lang, "qwen2.5:14b")
            return translated
        except Exception as e:
            logger.error(f"翻译出错: {str(e)}")
            return text  # 翻译失败时返回原文

    def translate_and_replace(self, image_path: str, output_path: str, 
                            source_lang: str = 'en', target_lang: str = 'zh',
                            min_confidence: float = 0.5) -> Optional[str]:
        """
        翻译图片中的文字并保持原始布局
        
        参数:
            image_path: 输入图片路径
            output_path: 输出图片路径
            source_lang: 源语言代码
            target_lang: 目标语言代码
            min_confidence: 最小置信度阈值
        """
        try:
            # 验证语言支持
            if source_lang not in self.supported_languages or target_lang not in self.supported_languages:
                raise ValueError(f"不支持的语言: {source_lang} -> {target_lang}")

            # 检查文件是否存在
            if not os.path.exists(image_path):
                raise FileNotFoundError(f"找不到输入图片: {image_path}")

            # 打开并处理图像
            image = Image.open(image_path).convert("RGB")
            draw = ImageDraw.Draw(image)
            
            # 初始化OCR
            reader = easyocr.Reader([source_lang])
            image_np = np.array(image)
            results = reader.readtext(image_np)

            logger.info(f"识别到 {len(results)} 个文本区域")

            for bbox, text, conf in results:
                if conf < min_confidence:
                    logger.warning(f"跳过低置信度文本: {text} (置信度: {conf})")
                    continue

                if not text.strip():
                    continue

                # 计算文本区域
                x1, y1 = map(int, bbox[0])
                x2, y2 = map(int, bbox[2])
                width = x2 - x1
                height = y2 - y1

                # 获取背景颜色
                bg_color = self.get_background_color(image, bbox)

                # 翻译文本
                translated = self.translate_text(text, source_lang, target_lang)
                logger.info(f"翻译: {text} -> {translated}")

                # 获取合适的字体和大小
                font_path = self.get_font_path(target_lang)
                font_size = self.calculate_font_size(translated, width, height, font_path)
                font = ImageFont.truetype(font_path, font_size)

                # 清除原始文本区域
                draw.rectangle([x1, y1, x2, y2], fill=bg_color)

                # 计算文本位置使其居中
                text_bbox = font.getbbox(translated)
                if text_bbox:
                    text_width = text_bbox[2] - text_bbox[0]
                    text_height = text_bbox[3] - text_bbox[1]
                    x = x1 + (width - text_width) // 2
                    y = y1 + (height - text_height) // 2
                else:
                    x, y = x1, y1

                # 绘制翻译后的文本
                draw.text((x, y), translated, fill='black', font=font)

            # 保存结果
            image.save(output_path)
            logger.info(f"翻译完成，已保存为 {output_path}")
            return output_path

        except Exception as e:
            logger.error(f"处理图片时出错: {str(e)}")
            return None

def main():
    translator = ImageTranslator()
    result = translator.translate_and_replace(
        "small_leftcorner.gif",
        "ThinkDesign_translated.jpg",
        source_lang='en',
        target_lang='zh'
    )
    
    if result:
        print(f"成功处理图片，输出保存在: {result}")
    else:
        print("处理图片时出错，请查看日志获取详细信息")

if __name__ == "__main__":
    main()
