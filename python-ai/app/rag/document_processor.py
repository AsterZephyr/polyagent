"""
Advanced Document Processing Module
多模态文档处理器：支持文本、PDF、图片、表格等多种格式
"""

import asyncio
import hashlib
import mimetypes
import re
from abc import ABC, abstractmethod
from typing import List, Dict, Any, Optional, Union, Tuple
from dataclasses import dataclass
from pathlib import Path
from datetime import datetime
import base64

import numpy as np
from PIL import Image
import fitz  # PyMuPDF
import docx
import pandas as pd
from bs4 import BeautifulSoup
import markdown

from app.core.logging import LoggerMixin
from app.core.exceptions import DocumentProcessingException
from app.rag.core import DocumentChunk, ChunkType

@dataclass
class ProcessingConfig:
    """文档处理配置"""
    
    # 通用分块参数
    chunk_size: int = 1000
    chunk_overlap: int = 200
    min_chunk_size: int = 100
    max_chunk_size: int = 4000
    
    # 文本处理
    preserve_formatting: bool = True
    extract_tables: bool = True
    extract_images: bool = True
    
    # 语义分块
    enable_semantic_chunking: bool = True
    sentence_window_size: int = 3
    similarity_threshold: float = 0.7
    
    # 质量控制
    min_quality_score: float = 0.3
    remove_duplicates: bool = True
    duplicate_threshold: float = 0.9
    
    # 元数据提取
    extract_entities: bool = True
    extract_keywords: bool = True
    extract_summary: bool = True

class BaseDocumentProcessor(ABC, LoggerMixin):
    """文档处理器基类"""
    
    def __init__(self, config: ProcessingConfig):
        self.config = config
    
    @abstractmethod
    async def process_document(
        self,
        content: Union[str, bytes],
        filename: str,
        metadata: Dict[str, Any] = None
    ) -> List[DocumentChunk]:
        """处理文档"""
        pass
    
    @abstractmethod
    def get_supported_types(self) -> List[str]:
        """获取支持的文件类型"""
        pass
    
    def _generate_chunk_id(self, doc_id: str, chunk_index: int) -> str:
        """生成块ID"""
        return f"{doc_id}_chunk_{chunk_index:04d}"
    
    def _calculate_quality_score(self, chunk: DocumentChunk) -> float:
        """计算块质量分数"""
        
        content = chunk.content.strip()
        
        if not content:
            return 0.0
        
        score = 1.0
        
        # 长度惩罚（太短或太长）
        length = len(content)
        if length < 50:
            score *= 0.5
        elif length > 3000:
            score *= 0.8
        
        # 内容质量评估
        
        # 1. 字符多样性
        unique_chars = len(set(content.lower()))
        char_diversity = unique_chars / len(content) if content else 0
        score *= min(char_diversity * 5, 1.0)
        
        # 2. 语言质量（简单启发式）
        words = content.split()
        if words:
            avg_word_length = np.mean([len(w) for w in words])
            if 3 <= avg_word_length <= 8:  # 正常范围
                score *= 1.0
            else:
                score *= 0.8
        
        # 3. 结构化程度
        if any(marker in content for marker in [".", "!", "?", ":", ";"]):
            score *= 1.1
        
        # 4. 特殊字符过多惩罚
        special_char_ratio = len(re.findall(r'[^\w\s\.,!?;:\-\(\)]', content)) / len(content) if content else 0
        if special_char_ratio > 0.1:
            score *= 0.7
        
        return min(score, 1.0)
    
    def _extract_keywords(self, content: str, max_keywords: int = 10) -> List[str]:
        """提取关键词（简化实现）"""
        
        # 移除停用词的简单实现
        stop_words = {
            'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 
            'of', 'with', 'by', 'is', 'are', 'was', 'were', 'be', 'been', 'being',
            'have', 'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could',
            'should', 'may', 'might', 'must', 'can', 'this', 'that', 'these', 'those'
        }
        
        # 简单的关键词提取
        words = re.findall(r'\b[a-zA-Z]{3,}\b', content.lower())
        word_freq = {}
        
        for word in words:
            if word not in stop_words:
                word_freq[word] = word_freq.get(word, 0) + 1
        
        # 按频率排序
        sorted_words = sorted(word_freq.items(), key=lambda x: x[1], reverse=True)
        
        return [word for word, _ in sorted_words[:max_keywords]]

class TextProcessor(BaseDocumentProcessor):
    """纯文本处理器"""
    
    def get_supported_types(self) -> List[str]:
        return ['.txt', '.md', '.rst', '.log']
    
    async def process_document(
        self,
        content: Union[str, bytes],
        filename: str,
        metadata: Dict[str, Any] = None
    ) -> List[DocumentChunk]:
        """处理文本文档"""
        
        if isinstance(content, bytes):
            try:
                text_content = content.decode('utf-8')
            except UnicodeDecodeError:
                text_content = content.decode('latin-1')
        else:
            text_content = content
        
        # 基础元数据
        doc_id = hashlib.md5(f"{filename}_{len(text_content)}".encode()).hexdigest()
        base_metadata = {
            "filename": filename,
            "doc_type": "text",
            "doc_id": doc_id,
            "total_chars": len(text_content),
            "processing_time": datetime.now().isoformat(),
            **(metadata or {})
        }
        
        # 分块处理
        if self.config.enable_semantic_chunking:
            chunks = await self._semantic_chunking(text_content, doc_id, base_metadata)
        else:
            chunks = await self._simple_chunking(text_content, doc_id, base_metadata)
        
        # 质量过滤
        high_quality_chunks = []
        for chunk in chunks:
            chunk.quality_score = self._calculate_quality_score(chunk)
            if chunk.quality_score >= self.config.min_quality_score:
                high_quality_chunks.append(chunk)
        
        self.logger.info(f"Processed {filename}: {len(chunks)} chunks, {len(high_quality_chunks)} high quality")
        
        return high_quality_chunks
    
    async def _simple_chunking(
        self,
        text: str,
        doc_id: str,
        base_metadata: Dict[str, Any]
    ) -> List[DocumentChunk]:
        """简单分块"""
        
        chunks = []
        chunk_size = self.config.chunk_size
        overlap = self.config.chunk_overlap
        
        start = 0
        chunk_index = 0
        
        while start < len(text):
            end = min(start + chunk_size, len(text))
            
            # 尝试在句子边界分割
            if end < len(text):
                # 寻找最近的句子结束
                sentence_end = text.rfind('.', start, end)
                if sentence_end > start + chunk_size // 2:
                    end = sentence_end + 1
            
            chunk_content = text[start:end].strip()
            
            if len(chunk_content) >= self.config.min_chunk_size:
                chunk = DocumentChunk(
                    id=self._generate_chunk_id(doc_id, chunk_index),
                    content=chunk_content,
                    chunk_type=ChunkType.PARAGRAPH,
                    source_doc_id=doc_id,
                    start_char=start,
                    end_char=end,
                    metadata={
                        **base_metadata,
                        "chunk_index": chunk_index,
                        "chunk_method": "simple"
                    },
                    relevance_keywords=self._extract_keywords(chunk_content)
                )
                
                chunks.append(chunk)
                chunk_index += 1
            
            start = max(start + chunk_size - overlap, end)
        
        return chunks
    
    async def _semantic_chunking(
        self,
        text: str,
        doc_id: str,
        base_metadata: Dict[str, Any]
    ) -> List[DocumentChunk]:
        """语义分块"""
        
        # 首先按段落分割
        paragraphs = text.split('\n\n')
        
        chunks = []
        current_chunk = ""
        chunk_index = 0
        start_char = 0
        
        for paragraph in paragraphs:
            paragraph = paragraph.strip()
            if not paragraph:
                continue
            
            # 检查是否应该开始新块
            potential_chunk = current_chunk + "\n\n" + paragraph if current_chunk else paragraph
            
            if len(potential_chunk) > self.config.max_chunk_size:
                # 保存当前块
                if current_chunk:
                    chunk = DocumentChunk(
                        id=self._generate_chunk_id(doc_id, chunk_index),
                        content=current_chunk.strip(),
                        chunk_type=ChunkType.PARAGRAPH,
                        source_doc_id=doc_id,
                        start_char=start_char,
                        end_char=start_char + len(current_chunk),
                        metadata={
                            **base_metadata,
                            "chunk_index": chunk_index,
                            "chunk_method": "semantic"
                        },
                        relevance_keywords=self._extract_keywords(current_chunk)
                    )
                    chunks.append(chunk)
                    chunk_index += 1
                
                # 开始新块
                start_char = start_char + len(current_chunk) + 2  # +2 for \n\n
                current_chunk = paragraph
            else:
                current_chunk = potential_chunk
        
        # 处理最后一个块
        if current_chunk and len(current_chunk.strip()) >= self.config.min_chunk_size:
            chunk = DocumentChunk(
                id=self._generate_chunk_id(doc_id, chunk_index),
                content=current_chunk.strip(),
                chunk_type=ChunkType.PARAGRAPH,
                source_doc_id=doc_id,
                start_char=start_char,
                end_char=start_char + len(current_chunk),
                metadata={
                    **base_metadata,
                    "chunk_index": chunk_index,
                    "chunk_method": "semantic"
                },
                relevance_keywords=self._extract_keywords(current_chunk)
            )
            chunks.append(chunk)
        
        return chunks

class PDFProcessor(BaseDocumentProcessor):
    """PDF处理器"""
    
    def get_supported_types(self) -> List[str]:
        return ['.pdf']
    
    async def process_document(
        self,
        content: Union[str, bytes],
        filename: str,
        metadata: Dict[str, Any] = None
    ) -> List[DocumentChunk]:
        """处理PDF文档"""
        
        if isinstance(content, str):
            # 如果是base64编码的字符串
            try:
                content = base64.b64decode(content)
            except:
                raise DocumentProcessingException(filename, "Invalid PDF content format")
        
        doc_id = hashlib.md5(f"{filename}_{len(content)}".encode()).hexdigest()
        
        try:
            # 使用PyMuPDF处理PDF
            pdf_document = fitz.open(stream=content, filetype="pdf")
            
            chunks = []
            chunk_index = 0
            
            for page_num in range(pdf_document.page_count):
                page = pdf_document[page_num]
                
                # 提取文本
                text_content = page.get_text()
                
                if text_content.strip():
                    # 文本分块处理
                    text_processor = TextProcessor(self.config)
                    page_chunks = await text_processor._simple_chunking(
                        text_content,
                        doc_id,
                        {
                            "filename": filename,
                            "doc_type": "pdf",
                            "doc_id": doc_id,
                            "page_number": page_num + 1,
                            "total_pages": pdf_document.page_count,
                            "processing_time": datetime.now().isoformat(),
                            **(metadata or {})
                        }
                    )
                    
                    # 更新块信息
                    for chunk in page_chunks:
                        chunk.id = self._generate_chunk_id(doc_id, chunk_index)
                        chunk.page_number = page_num + 1
                        chunk.metadata["chunk_index"] = chunk_index
                        chunk.quality_score = self._calculate_quality_score(chunk)
                        
                        if chunk.quality_score >= self.config.min_quality_score:
                            chunks.append(chunk)
                        
                        chunk_index += 1
                
                # 提取表格（如果启用）
                if self.config.extract_tables:
                    table_chunks = self._extract_tables_from_page(page, doc_id, page_num + 1, chunk_index, metadata)
                    chunks.extend(table_chunks)
                    chunk_index += len(table_chunks)
                
                # 提取图片（如果启用）
                if self.config.extract_images:
                    image_chunks = self._extract_images_from_page(page, doc_id, page_num + 1, chunk_index, metadata)
                    chunks.extend(image_chunks)
                    chunk_index += len(image_chunks)
            
            pdf_document.close()
            
            self.logger.info(f"Processed PDF {filename}: {len(chunks)} chunks from {pdf_document.page_count} pages")
            
            return chunks
            
        except Exception as e:
            raise DocumentProcessingException(filename, f"PDF processing failed: {str(e)}")
    
    def _extract_tables_from_page(
        self,
        page,
        doc_id: str,
        page_num: int,
        start_index: int,
        base_metadata: Dict[str, Any]
    ) -> List[DocumentChunk]:
        """从页面提取表格"""
        
        chunks = []
        
        try:
            # 简化的表格检测（实际应该使用更复杂的算法）
            tables = page.find_tables()
            
            for i, table in enumerate(tables):
                try:
                    # 提取表格数据
                    table_data = table.extract()
                    
                    # 转换为文本格式
                    table_text = self._table_to_text(table_data)
                    
                    if table_text.strip():
                        chunk = DocumentChunk(
                            id=self._generate_chunk_id(doc_id, start_index + i),
                            content=table_text,
                            chunk_type=ChunkType.TABLE,
                            source_doc_id=doc_id,
                            page_number=page_num,
                            metadata={
                                **base_metadata,
                                "chunk_index": start_index + i,
                                "table_index": i,
                                "chunk_method": "table_extraction"
                            }
                        )
                        
                        chunk.quality_score = self._calculate_quality_score(chunk)
                        if chunk.quality_score >= self.config.min_quality_score:
                            chunks.append(chunk)
                
                except Exception as e:
                    self.logger.warning(f"Failed to extract table {i} from page {page_num}: {str(e)}")
        
        except Exception as e:
            self.logger.warning(f"Table extraction failed for page {page_num}: {str(e)}")
        
        return chunks
    
    def _extract_images_from_page(
        self,
        page,
        doc_id: str,
        page_num: int,
        start_index: int,
        base_metadata: Dict[str, Any]
    ) -> List[DocumentChunk]:
        """从页面提取图片"""
        
        chunks = []
        
        try:
            image_list = page.get_images()
            
            for i, img in enumerate(image_list):
                try:
                    # 获取图片
                    xref = img[0]
                    pix = fitz.Pixmap(page.parent, xref)
                    
                    if pix.n - pix.alpha < 4:  # 确保不是CMYK
                        # 这里应该使用OCR或图像描述模型来生成文本
                        # 暂时使用占位符
                        image_description = f"Image {i+1} on page {page_num} (dimensions: {pix.width}x{pix.height})"
                        
                        chunk = DocumentChunk(
                            id=self._generate_chunk_id(doc_id, start_index + i),
                            content=image_description,
                            chunk_type=ChunkType.IMAGE_CAPTION,
                            source_doc_id=doc_id,
                            page_number=page_num,
                            metadata={
                                **base_metadata,
                                "chunk_index": start_index + i,
                                "image_index": i,
                                "image_width": pix.width,
                                "image_height": pix.height,
                                "chunk_method": "image_extraction"
                            }
                        )
                        
                        chunks.append(chunk)
                    
                    pix = None  # 释放内存
                
                except Exception as e:
                    self.logger.warning(f"Failed to extract image {i} from page {page_num}: {str(e)}")
        
        except Exception as e:
            self.logger.warning(f"Image extraction failed for page {page_num}: {str(e)}")
        
        return chunks
    
    def _table_to_text(self, table_data: List[List[str]]) -> str:
        """将表格数据转换为文本"""
        
        if not table_data:
            return ""
        
        # 简单的表格文本化
        lines = []
        for row in table_data:
            if row:
                clean_row = [str(cell).strip() if cell else "" for cell in row]
                lines.append(" | ".join(clean_row))
        
        return "\n".join(lines)

class DocumentProcessorFactory(LoggerMixin):
    """文档处理器工厂"""
    
    def __init__(self, config: ProcessingConfig):
        self.config = config
        self.processors: Dict[str, BaseDocumentProcessor] = {}
        
        # 注册默认处理器
        self._register_default_processors()
    
    def _register_default_processors(self):
        """注册默认处理器"""
        
        # 文本处理器
        text_processor = TextProcessor(self.config)
        for ext in text_processor.get_supported_types():
            self.processors[ext] = text_processor
        
        # PDF处理器
        pdf_processor = PDFProcessor(self.config)
        for ext in pdf_processor.get_supported_types():
            self.processors[ext] = pdf_processor
    
    def register_processor(self, extensions: List[str], processor: BaseDocumentProcessor):
        """注册处理器"""
        for ext in extensions:
            self.processors[ext] = processor
            self.logger.info(f"Registered processor for {ext}")
    
    async def process_document(
        self,
        content: Union[str, bytes],
        filename: str,
        metadata: Dict[str, Any] = None
    ) -> List[DocumentChunk]:
        """处理文档"""
        
        # 获取文件扩展名
        ext = Path(filename).suffix.lower()
        
        if ext not in self.processors:
            # 尝试通过MIME类型推断
            mime_type, _ = mimetypes.guess_type(filename)
            if mime_type:
                if mime_type.startswith('text/'):
                    ext = '.txt'
                elif mime_type == 'application/pdf':
                    ext = '.pdf'
        
        if ext not in self.processors:
            raise DocumentProcessingException(
                filename, 
                f"Unsupported file type: {ext}. Supported types: {list(self.processors.keys())}"
            )
        
        processor = self.processors[ext]
        
        try:
            chunks = await processor.process_document(content, filename, metadata)
            
            # 去重处理（如果启用）
            if self.config.remove_duplicates:
                chunks = self._remove_duplicates(chunks)
            
            return chunks
            
        except Exception as e:
            self.logger.error(f"Document processing failed for {filename}: {str(e)}")
            raise DocumentProcessingException(filename, str(e))
    
    def _remove_duplicates(self, chunks: List[DocumentChunk]) -> List[DocumentChunk]:
        """去除重复块"""
        
        if not chunks:
            return chunks
        
        unique_chunks = []
        seen_hashes = set()
        
        for chunk in chunks:
            # 生成内容哈希
            content_hash = hashlib.md5(chunk.content.encode()).hexdigest()
            
            if content_hash not in seen_hashes:
                seen_hashes.add(content_hash)
                unique_chunks.append(chunk)
            else:
                self.logger.debug(f"Removed duplicate chunk: {chunk.id}")
        
        self.logger.info(f"Deduplication: {len(chunks)} -> {len(unique_chunks)} chunks")
        
        return unique_chunks
    
    def get_supported_types(self) -> List[str]:
        """获取支持的文件类型"""
        return list(self.processors.keys())