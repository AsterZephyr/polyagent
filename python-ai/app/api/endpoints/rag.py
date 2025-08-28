"""
RAG端点
"""

from fastapi import APIRouter, HTTPException, Depends
from app.api.models import RAGQuery, RAGResult, DocumentUploadRequest, DocumentUploadResponse
from app.services.rag_service import RAGService
from app.core.logging import get_logger

router = APIRouter()
logger = get_logger("rag")

async def get_rag_service() -> RAGService:
    """获取RAG服务实例"""
    from main import rag_service
    if not rag_service:
        raise HTTPException(status_code=503, detail="RAG service not available")
    return rag_service

@router.post("/query", response_model=RAGResult)
async def query_documents(
    request: RAGQuery,
    rag_service: RAGService = Depends(get_rag_service)
):
    """查询文档"""
    
    try:
        result = await rag_service.query(
            user_id=request.user_id,
            query=request.query,
            top_k=request.top_k,
            filters=request.filters
        )
        
        return RAGResult(**result)
        
    except Exception as e:
        logger.error(f"RAG query failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/upload", response_model=DocumentUploadResponse)
async def upload_document(
    request: DocumentUploadRequest,
    rag_service: RAGService = Depends(get_rag_service)
):
    """上传文档"""
    
    try:
        result = await rag_service.upload_document(
            user_id=request.user_id,
            filename=request.filename,
            content=request.content
        )
        
        return DocumentUploadResponse(**result)
        
    except Exception as e:
        logger.error(f"Document upload failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))