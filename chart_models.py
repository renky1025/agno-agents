from typing import List, Union, Dict, Any, Optional
from pydantic import BaseModel


class Dataset(BaseModel):
    type: Optional[str] = None
    fill: Optional[bool] = None
    label: Optional[str] = None
    data: Any
    backgroundColor: Optional[Union[str, List[str]]] = None
    borderColor: Optional[Union[str, List[str]]] = None


class ChartData(BaseModel):
    labels: Optional[List[str]] = None
    datasets: List[Dataset]


class TitleOptions(BaseModel):
    display: Optional[bool] = None
    text: Optional[str] = None


class ChartOptions(BaseModel):
    title: Optional[TitleOptions] = None


class ChartConfig(BaseModel):
    type: str
    data: ChartData
    options: Optional[ChartOptions] = None 