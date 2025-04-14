import requests

class TranslatorService:
    def __init__(self):
        self.base_url = "http://10.100.1.1:11434/api"
    
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
            
        prompt = f"Translate the following {source_lang} text to {target_lang}. C++ class name and method name are not translated, and the program code does not need to be translated as is. Thinkdesign is the brand name and does not need to be translated. Output the translation only without any explanations or additional context:\n{text}"
        
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

# if __name__ == "__main__":
#     translator = TranslatorService()
#     print(translator.get_available_models())
#     print(translator.translate("Hello, world!", "en", "zh", "gemma3:1b"))

