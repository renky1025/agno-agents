import requests
from typing import Dict, Optional, List, Tuple
from difflib import SequenceMatcher
import re

class TranslatorService:
    def __init__(self):
        self.base_url = "http://10.100.1.1:11434/api"
        self.custom_dict: Dict[str, str] = {}
        self.similarity_threshold = 0.4  # 相似度阈值
        self.max_dict_entries = 20  # prompt中最大字典条目数
    
    def set_custom_dictionary(self, dictionary: Dict[str, str]) -> None:
        """设置自定义翻译字典
        
        Args:
            dictionary: 包含源语言词汇到目标语言翻译的字典
        """
        self.custom_dict = dictionary

    def get_custom_dictionary(self) -> Dict[str, str]:
        """获取当前的自定义翻译字典"""
        return self.custom_dict

    def clear_custom_dictionary(self) -> None:
        """清空自定义翻译字典"""
        self.custom_dict = {}
    
    def set_similarity_threshold(self, threshold: float) -> None:
        """设置相似度阈值
        
        Args:
            threshold: 0到1之间的数值，越大表示匹配越严格
        """
        if 0 <= threshold <= 1:
            self.similarity_threshold = threshold
    
    def set_max_dict_entries(self, max_entries: int) -> None:
        """设置prompt中最大字典条目数
        
        Args:
            max_entries: 最大条目数
        """
        if max_entries > 0:
            self.max_dict_entries = max_entries

    def _calculate_similarity(self, str1: str, str2: str) -> float:
        """计算两个字符串的相似度"""
        return SequenceMatcher(None, str1.lower(), str2.lower()).ratio()

    def _extract_words(self, text: str) -> List[str]:
        """从文本中提取单词"""
        # 使用正则表达式匹配单词，包括可能的专业术语
        words = re.findall(r'\b\w+\b', text)
        # 添加多词组合（最多3个词的组合）
        phrases = []
        for i in range(len(words)):
            for j in range(1, 4):
                if i + j <= len(words):
                    phrase = ' '.join(words[i:i+j])
                    phrases.append(phrase)
        return list(set(phrases))

    def _find_relevant_dict_entries(self, text: str) -> Dict[str, str]:
        """找出与给定文本相关的字典条目
        
        Args:
            text: 待翻译的文本
        
        Returns:
            相关的字典条目
        """
        if not self.custom_dict:
            return {}

        # 从文本中提取单词和短语
        text_words = self._extract_words(text)
        
        # 计算每个字典条目与文本中单词的相似度
        relevant_entries: List[Tuple[str, str, float]] = []
        for dict_key, dict_value in self.custom_dict.items():
            max_similarity = 0
            for word in text_words:
                similarity = self._calculate_similarity(dict_key, word)
                max_similarity = max(max_similarity, similarity)
            
            if max_similarity >= self.similarity_threshold:
                relevant_entries.append((dict_key, dict_value, max_similarity))
        
        # 按相似度排序并限制条目数量
        relevant_entries.sort(key=lambda x: x[2], reverse=True)
        relevant_entries = relevant_entries[:self.max_dict_entries]
        
        return {entry[0]: entry[1] for entry in relevant_entries}

    def get_available_models(self) -> list:
        """获取可用的模型列表"""
        try:
            response = requests.get(f"{self.base_url}/tags")
            if response.status_code == 200:
                return response.json().get('models', [])
            return []
        except Exception:
            return []
    
    def translate(self, 
                  text: str, 
                  source_lang: str, 
                  target_lang: str, 
                  model: str,
                  temperature: float = 0.7,
                  top_p: float = 0.9) -> str:
        """执行翻译请求"""
        if not text or not model:
            return ""
        
        # 获取相关的字典条目
        relevant_dict = self._find_relevant_dict_entries(text)
        
        # 构建自定义字典提示
        custom_dict_prompt = ""
        if relevant_dict:
            custom_dict_prompt = "Please use the following custom dictionary for translation:\n"
            for source, target in relevant_dict.items():
                custom_dict_prompt += f"- {source} → {target}\n"
            custom_dict_prompt += "\n"
            
        prompt = f"""Translate the following {source_lang} text to {target_lang}.
{custom_dict_prompt}Rules:
1. C++ class name and method name are not translated
2. Program code does not need to be translated
3. Thinkdesign is the brand name and does not need to be translated
4. Strictly follow the custom dictionary translations when those terms appear
5. Output the translation only without any explanations or additional context

Text to translate:
{text}"""
        
        try:
            response = requests.post(
                f"{self.base_url}/generate",
                json={
                    "model": model,
                    "prompt": prompt,
                    "stream": False,
                    "options": {
                        "temperature": temperature,
                        "top_p": top_p
                    }
                }
            )
            
            if response.status_code == 200:
                return response.json().get('response', '')
            return f"翻译失败: HTTP {response.status_code}"
        except Exception as e:
            return f"翻译失败: {str(e)}"

# 使用示例
if __name__ == "__main__":
    translator = TranslatorService()
    
    # 设置自定义字典
    custom_dict = {
        "Hello": "您好",
        "world": "世界",
        "technical documentation": "技术文档",
        "software development": "软件开发",
        "artificial intelligence": "人工智能"
    }
    translator.set_custom_dictionary(custom_dict)
    
    # 设置相似度阈值（可选）
    translator.set_similarity_threshold(0.4)
    
    # 设置最大字典条目数（可选）
    translator.set_max_dict_entries(20)
    
    # 测试翻译
    text = "Hello, this is a technical documentation about software development and artificial intelligence."
    print(translator.translate(text, "en", "zh", "gemma3:4b"))

