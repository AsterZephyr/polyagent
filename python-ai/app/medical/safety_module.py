"""
Medical Safety Module
Handles medical AI safety, hallucination detection, and compliance
"""

import asyncio
import re
from typing import Dict, List, Any, Optional, Tuple, Set
from dataclasses import dataclass
from enum import Enum
import json
from datetime import datetime

from ..core.logging import LoggerMixin
from ..core.base_exceptions import ComponentExecutionException

class RiskLevel(Enum):
    """Medical information risk levels"""
    LOW = "low"           # General health information
    MEDIUM = "medium"     # Specific symptoms/conditions
    HIGH = "high"         # Treatment recommendations
    CRITICAL = "critical" # Emergency situations

class VerificationStatus(Enum):
    """Verification status for medical claims"""
    VERIFIED = "verified"
    UNCERTAIN = "uncertain"
    CONTRADICTED = "contradicted"
    INSUFFICIENT_DATA = "insufficient_data"

@dataclass
class MedicalClaim:
    """Medical claim for verification"""
    text: str
    category: str  # symptom, treatment, diagnosis, etc.
    risk_level: RiskLevel
    confidence_score: float = 0.0
    sources: List[str] = None
    verification_status: VerificationStatus = VerificationStatus.INSUFFICIENT_DATA

@dataclass
class SafetyCheckResult:
    """Result of medical safety check"""
    is_safe: bool
    risk_level: RiskLevel
    issues_found: List[str]
    recommendations: List[str]
    requires_disclaimer: bool
    blocked_content: List[str]
    confidence_score: float

class MedicalSafetyChecker(LoggerMixin):
    """Medical AI safety checker and hallucination detector"""
    
    def __init__(self):
        super().__init__()
        self.setup_safety_rules()
        self.setup_medical_knowledge_base()
    
    def setup_safety_rules(self):
        """Setup medical safety rules"""
        
        # Prohibited content patterns
        self.prohibited_patterns = [
            r"诊断为|确诊为|我诊断你",  # Diagnosis claims
            r"建议立即服用|推荐药物|服用.*毫克",  # Medication recommendations
            r"不需要看医生|无需就医",  # Discouraging medical consultation
            r"这是癌症|你患有.*病",  # Direct disease attribution
            r"立即手术|需要手术",  # Surgical recommendations
        ]
        
        # High-risk keywords requiring extra caution
        self.high_risk_keywords = {
            'emergency': ['心脏病', '中风', '昏迷', '大出血', '呼吸困难', '胸痛'],
            'diagnosis': ['癌症', '肿瘤', '糖尿病', '高血压', '抑郁症'],
            'medication': ['抗生素', '胰岛素', '激素', '化疗', '手术'],
            'children': ['婴儿', '儿童', '新生儿', '怀孕', '孕妇']
        }
        
        # Required disclaimers by risk level
        self.disclaimers = {
            RiskLevel.LOW: "此信息仅供参考，不能替代专业医疗建议。",
            RiskLevel.MEDIUM: "⚠️ 此信息仅供参考，如有症状请咨询医疗专业人员。",
            RiskLevel.HIGH: "⚠️ 重要：此信息不能替代医疗诊断，请立即咨询合格医生。",
            RiskLevel.CRITICAL: "🚨 紧急提醒：如遇紧急情况请立即就医或拨打急救电话！"
        }
    
    def setup_medical_knowledge_base(self):
        """Setup medical knowledge base for verification"""
        # This would connect to authoritative medical databases
        self.knowledge_sources = {
            "symptom_database": "医学症状数据库",
            "drug_database": "药物安全数据库", 
            "treatment_guidelines": "医疗指南数据库",
            "medical_literature": "医学文献数据库"
        }
        
        # Common medical fact patterns for verification
        self.fact_patterns = {
            "vital_signs": {
                "血压正常范围": {"systolic": (90, 140), "diastolic": (60, 90)},
                "正常体温": (36.1, 37.2),
                "正常心率": (60, 100)
            },
            "emergency_signs": [
                "胸痛", "呼吸困难", "昏迷", "大量出血", "持续高烧"
            ]
        }
    
    async def check_response_safety(self, 
                                  response: str, 
                                  query: str = "",
                                  context: Dict[str, Any] = None) -> SafetyCheckResult:
        """Comprehensive safety check for medical response"""
        
        issues = []
        recommendations = []
        blocked_content = []
        risk_level = RiskLevel.LOW
        
        # 1. Check for prohibited patterns
        prohibited_issues = self._check_prohibited_patterns(response)
        if prohibited_issues:
            issues.extend(prohibited_issues)
            blocked_content.extend(prohibited_issues)
            risk_level = RiskLevel.HIGH
        
        # 2. Assess risk level based on content
        content_risk = self._assess_content_risk(response)
        risk_level = max(risk_level, content_risk)
        
        # 3. Check for medical claims that need verification
        unverified_claims = await self._check_medical_claims(response)
        if unverified_claims:
            issues.append(f"包含 {len(unverified_claims)} 个未验证的医学声明")
            recommendations.append("建议验证医学声明的准确性")
        
        # 4. Check for hallucination indicators
        hallucination_risk = await self._detect_hallucination_risk(response, query)
        if hallucination_risk > 0.5:
            issues.append("检测到可能的AI幻觉内容")
            recommendations.append("需要额外验证内容准确性")
            risk_level = max(risk_level, RiskLevel.MEDIUM)
        
        # 5. Check completeness and accuracy
        completeness_score = self._check_response_completeness(response, query)
        
        # 6. Final safety assessment
        is_safe = (len(blocked_content) == 0 and 
                  risk_level != RiskLevel.CRITICAL and
                  hallucination_risk < 0.7)
        
        # Add risk-appropriate recommendations
        if risk_level >= RiskLevel.MEDIUM:
            recommendations.append("添加医疗免责声明")
        
        if risk_level >= RiskLevel.HIGH:
            recommendations.append("建议用户咨询医疗专业人员")
        
        confidence_score = max(0.0, 1.0 - hallucination_risk - (len(unverified_claims) * 0.1))
        
        return SafetyCheckResult(
            is_safe=is_safe,
            risk_level=risk_level,
            issues_found=issues,
            recommendations=recommendations,
            requires_disclaimer=risk_level >= RiskLevel.MEDIUM,
            blocked_content=blocked_content,
            confidence_score=confidence_score
        )
    
    def _check_prohibited_patterns(self, text: str) -> List[str]:
        """Check for prohibited medical content patterns"""
        violations = []
        
        for pattern in self.prohibited_patterns:
            matches = re.findall(pattern, text, re.IGNORECASE)
            if matches:
                violations.append(f"包含禁止的医疗声明: {matches[0]}")
        
        return violations
    
    def _assess_content_risk(self, text: str) -> RiskLevel:
        """Assess risk level based on content analysis"""
        text_lower = text.lower()
        
        # Check for emergency keywords
        if any(keyword in text_lower for keyword in self.high_risk_keywords['emergency']):
            return RiskLevel.CRITICAL
        
        # Check for diagnostic content
        if any(keyword in text_lower for keyword in self.high_risk_keywords['diagnosis']):
            return RiskLevel.HIGH
        
        # Check for medication content
        if any(keyword in text_lower for keyword in self.high_risk_keywords['medication']):
            return RiskLevel.HIGH
        
        # Check for children/pregnancy content
        if any(keyword in text_lower for keyword in self.high_risk_keywords['children']):
            return RiskLevel.MEDIUM
        
        return RiskLevel.LOW
    
    async def _check_medical_claims(self, text: str) -> List[MedicalClaim]:
        """Extract and verify medical claims"""
        claims = []
        
        # Extract potential medical claims using patterns
        claim_patterns = [
            r"正常.*范围.*[0-9]+",  # Normal ranges
            r"症状包括.*",          # Symptoms
            r"治疗方法.*",          # Treatments
            r".*会导致.*"           # Causation claims
        ]
        
        for pattern in claim_patterns:
            matches = re.findall(pattern, text)
            for match in matches:
                claim = MedicalClaim(
                    text=match,
                    category="general",
                    risk_level=self._assess_claim_risk(match)
                )
                
                # Attempt to verify claim
                claim.verification_status = await self._verify_medical_claim(claim)
                claims.append(claim)
        
        return [claim for claim in claims 
                if claim.verification_status in [VerificationStatus.UNCERTAIN, 
                                               VerificationStatus.CONTRADICTED]]
    
    async def _verify_medical_claim(self, claim: MedicalClaim) -> VerificationStatus:
        """Verify medical claim against knowledge base"""
        # This would implement actual verification against medical databases
        
        # For now, implement basic verification logic
        claim_text = claim.text.lower()
        
        # Check vital signs claims
        if "血压" in claim_text and "正常" in claim_text:
            # Extract numbers and check against known ranges
            numbers = re.findall(r'\d+', claim.text)
            if len(numbers) >= 2:
                systolic, diastolic = int(numbers[0]), int(numbers[1])
                normal_range = self.fact_patterns["vital_signs"]["血压正常范围"]
                
                if (normal_range["systolic"][0] <= systolic <= normal_range["systolic"][1] and
                    normal_range["diastolic"][0] <= diastolic <= normal_range["diastolic"][1]):
                    return VerificationStatus.VERIFIED
                else:
                    return VerificationStatus.CONTRADICTED
        
        # Default to uncertain for complex claims
        return VerificationStatus.UNCERTAIN
    
    def _assess_claim_risk(self, claim_text: str) -> RiskLevel:
        """Assess risk level of a medical claim"""
        if any(keyword in claim_text.lower() for keyword in self.high_risk_keywords['emergency']):
            return RiskLevel.CRITICAL
        elif any(keyword in claim_text.lower() for keyword in self.high_risk_keywords['diagnosis']):
            return RiskLevel.HIGH
        elif any(keyword in claim_text.lower() for keyword in self.high_risk_keywords['medication']):
            return RiskLevel.HIGH
        else:
            return RiskLevel.MEDIUM
    
    async def _detect_hallucination_risk(self, response: str, query: str) -> float:
        """Detect potential AI hallucination in medical context"""
        risk_score = 0.0
        
        # Check for overly specific medical claims without qualification
        specific_patterns = [
            r'\d+%的.*患者',      # Specific percentages
            r'研究表明.*\d+',     # Studies with specific numbers
            r'医生建议.*',        # Doctor recommendations
            r'根据.*医院.*'       # Hospital references
        ]
        
        for pattern in specific_patterns:
            if re.search(pattern, response):
                risk_score += 0.2
        
        # Check for contradictory information within response
        if self._contains_contradictions(response):
            risk_score += 0.3
        
        # Check for unrealistic claims
        unrealistic_patterns = [
            r'100%.*治愈',      # 100% cure claims
            r'绝对.*安全',      # Absolute safety claims
            r'永远不会.*',      # Never claims
            r'立即.*治好'       # Immediate cure claims
        ]
        
        for pattern in unrealistic_patterns:
            if re.search(pattern, response, re.IGNORECASE):
                risk_score += 0.4
        
        # Normalize score
        return min(risk_score, 1.0)
    
    def _contains_contradictions(self, text: str) -> bool:
        """Simple contradiction detection"""
        # Look for contradictory statements
        contradictory_pairs = [
            (['安全', '无害'], ['危险', '有害']),
            (['推荐', '建议'], ['不推荐', '不建议']),
            (['正常', '健康'], ['异常', '不健康'])
        ]
        
        text_lower = text.lower()
        for positive, negative in contradictory_pairs:
            has_positive = any(word in text_lower for word in positive)
            has_negative = any(word in text_lower for word in negative)
            
            if has_positive and has_negative:
                return True
        
        return False
    
    def _check_response_completeness(self, response: str, query: str) -> float:
        """Check if response adequately addresses the query"""
        # Simple completeness scoring
        query_keywords = set(re.findall(r'\w+', query.lower()))
        response_keywords = set(re.findall(r'\w+', response.lower()))
        
        overlap = len(query_keywords & response_keywords)
        total_query_keywords = len(query_keywords)
        
        if total_query_keywords == 0:
            return 1.0
        
        return overlap / total_query_keywords
    
    async def apply_safety_measures(self, 
                                   response: str, 
                                   safety_result: SafetyCheckResult) -> str:
        """Apply safety measures to response"""
        
        if not safety_result.is_safe:
            # Block unsafe content
            if safety_result.blocked_content:
                return "抱歉，我无法提供可能不安全的医疗建议。请咨询合格的医疗专业人员。"
        
        modified_response = response
        
        # Add disclaimer based on risk level
        if safety_result.requires_disclaimer:
            disclaimer = self.disclaimers[safety_result.risk_level]
            modified_response = f"{response}\n\n{disclaimer}"
        
        # Add verification notice for uncertain claims
        if safety_result.confidence_score < 0.7:
            verification_notice = "\n\n📋 请注意：上述信息的准确性需要进一步验证，建议咨询医疗专业人员。"
            modified_response += verification_notice
        
        # Add emergency notice for high-risk content
        if safety_result.risk_level == RiskLevel.CRITICAL:
            emergency_notice = "\n\n🚨 如果您正在经历医疗紧急情况，请立即拨打急救电话或前往最近的医院！"
            modified_response += emergency_notice
        
        return modified_response
    
    async def generate_safety_report(self, 
                                   query: str, 
                                   response: str, 
                                   safety_result: SafetyCheckResult) -> Dict[str, Any]:
        """Generate detailed safety report for logging"""
        
        return {
            "timestamp": datetime.now().isoformat(),
            "query": query,
            "response_length": len(response),
            "safety_assessment": {
                "is_safe": safety_result.is_safe,
                "risk_level": safety_result.risk_level.value,
                "confidence_score": safety_result.confidence_score,
                "issues_count": len(safety_result.issues_found),
                "requires_disclaimer": safety_result.requires_disclaimer
            },
            "issues_found": safety_result.issues_found,
            "recommendations": safety_result.recommendations,
            "blocked_content": safety_result.blocked_content,
            "verification_status": "pending_expert_review" if safety_result.confidence_score < 0.5 else "auto_approved"
        }

class MedicalFactVerifier(LoggerMixin):
    """Dedicated fact verification for medical claims"""
    
    def __init__(self):
        super().__init__()
        self.authoritative_sources = {
            "who": "World Health Organization",
            "cdc": "Centers for Disease Control", 
            "mayo": "Mayo Clinic",
            "pubmed": "PubMed Medical Literature"
        }
    
    async def verify_medical_fact(self, 
                                 claim: str, 
                                 sources: List[str] = None) -> Dict[str, Any]:
        """Verify medical fact against authoritative sources"""
        
        verification_result = {
            "claim": claim,
            "verified": False,
            "confidence": 0.0,
            "sources_checked": [],
            "supporting_evidence": [],
            "contradicting_evidence": [],
            "recommendation": "expert_verification_required"
        }
        
        # This would implement actual fact-checking against medical databases
        # For now, implement basic pattern matching
        
        if self._is_basic_fact(claim):
            verification_result.update({
                "verified": True,
                "confidence": 0.8,
                "recommendation": "likely_accurate"
            })
        
        return verification_result
    
    def _is_basic_fact(self, claim: str) -> bool:
        """Check if claim is a basic medical fact"""
        basic_facts = [
            "正常体温", "血压范围", "心率正常", "多喝水", "充足睡眠", "均衡饮食"
        ]
        
        return any(fact in claim for fact in basic_facts)