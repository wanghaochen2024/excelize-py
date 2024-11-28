"""Copyright 2024 The excelize Authors. All rights reserved. Use of this source
code is governed by a BSD-style license that can be found in the LICENSE file.

Package excelize-py is a Python port of Go Excelize library, providing a set of
functions that allow you to write and read from XLAM / XLSM / XLSX / XLTM / XLTX
files. Supports reading and writing spreadsheet documents generated by Microsoft
Excel™ 2007 and later. Supports complex components by high compatibility, and
provided streaming API for generating or reading data from a worksheet with huge
amounts of data. This library needs Python version 3.9 or later.
"""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional


class CultureName(IntEnum):
    CultureNameUnknown = 0
    CultureNameEnUS = 1
    CultureNameJaJP = 2
    CultureNameKoKR = 3
    CultureNameZhCN = 4
    CultureNameZhTW = 5


@dataclass
class Interface:
    type: int = 0
    integer: int = 0
    string: str = ""
    float64: float = 0
    boolean: bool = False


@dataclass
class Options:
    max_calc_iterations: int = 0
    password: str = ""
    raw_cell_value: bool = False
    unzip_size_limit: int = 0
    unzip_xml_size_limit: int = 0
    short_date_pattern: str = ""
    long_date_pattern: str = ""
    long_time_pattern: str = ""
    culture_info: CultureName = CultureName.CultureNameUnknown


@dataclass
class Border:
    type: str = ""
    color: str = ""
    style: int = 0


@dataclass
class Fill:
    type: str = ""
    pattern: int = 0
    color: Optional[list[str]] = None
    shading: int = 0


@dataclass
class Font:
    bold: bool = False
    italic: bool = False
    underline: str = ""
    family: str = ""
    size: float = 0
    strike: bool = False
    color: str = ""
    color_indexed: int = 0
    color_theme: Optional[int] = None
    color_tint: float = 0
    vert_align: str = ""


@dataclass
class Alignment:
    horizontal: str = ""
    indent: int = 0
    justify_last_line: bool = False
    reading_order: int = 0
    relative_indent: int = 0
    shrink_to_fit: bool = False
    text_rotation: int = 0
    vertical: str = ""
    wrap_text: bool = False


@dataclass
class Protection:
    hidden: bool = False
    locked: bool = False


@dataclass
class Style:
    border: Optional[list[Border]] = None
    fill: Fill = Fill
    font: Optional[Font] = None
    alignment: Optional[Alignment] = None
    protection: Optional[Protection] = None
    num_fmt: int = 0
    decimal_places: Optional[int] = None
    custom_num_fmt: Optional[str] = None
    neg_red: bool = False
