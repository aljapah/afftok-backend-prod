"""
AffTok Server-to-Server Integration - Python/FastAPI Example

This example shows how to send postbacks/conversions from your server
to AffTok using the Server-to-Server API.
"""

import os
import time
import hmac
import hashlib
import random
import string
from typing import Optional, Dict, Any, List
import httpx
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

# Configuration
class Config:
    API_KEY = os.getenv("AFFTOK_API_KEY", "your_api_key")
    ADVERTISER_ID = os.getenv("AFFTOK_ADVERTISER_ID", "your_advertiser_id")
    BASE_URL = os.getenv("AFFTOK_BASE_URL", "https://api.afftok.com")
    TIMEOUT = 30


# Pydantic models
class ConversionParams(BaseModel):
    offer_id: str
    transaction_id: str
    click_id: Optional[str] = None
    amount: Optional[float] = None
    currency: str = "USD"
    status: str = "approved"
    custom_params: Optional[Dict[str, Any]] = None


class ClickParams(BaseModel):
    offer_id: str
    tracking_code: Optional[str] = None
    sub_id_1: Optional[str] = None
    sub_id_2: Optional[str] = None
    sub_id_3: Optional[str] = None
    ip: Optional[str] = None
    user_agent: Optional[str] = None
    custom_params: Optional[Dict[str, Any]] = None


class AfftokTracker:
    """AffTok Server-to-Server Tracker"""
    
    def __init__(
        self,
        api_key: str = Config.API_KEY,
        advertiser_id: str = Config.ADVERTISER_ID,
        base_url: str = Config.BASE_URL
    ):
        self.api_key = api_key
        self.advertiser_id = advertiser_id
        self.base_url = base_url
        self.client = httpx.AsyncClient(timeout=Config.TIMEOUT)
    
    def _generate_signature(self, timestamp: int, nonce: str) -> str:
        """Generate HMAC-SHA256 signature"""
        data_to_sign = f"{self.api_key}|{self.advertiser_id}|{timestamp}|{nonce}"
        return hmac.new(
            self.api_key.encode(),
            data_to_sign.encode(),
            hashlib.sha256
        ).hexdigest()
    
    def _generate_nonce(self, length: int = 32) -> str:
        """Generate random nonce"""
        chars = string.ascii_letters + string.digits
        return ''.join(random.choice(chars) for _ in range(length))
    
    async def send_postback(self, params: ConversionParams) -> Dict[str, Any]:
        """
        Send a postback/conversion to AffTok
        
        Args:
            params: Conversion parameters
            
        Returns:
            Result dictionary with success status
        """
        timestamp = int(time.time() * 1000)
        nonce = self._generate_nonce()
        signature = self._generate_signature(timestamp, nonce)
        
        payload = {
            "api_key": self.api_key,
            "advertiser_id": self.advertiser_id,
            "offer_id": params.offer_id,
            "transaction_id": params.transaction_id,
            "status": params.status,
            "currency": params.currency,
            "timestamp": timestamp,
            "nonce": nonce,
            "signature": signature,
        }
        
        # Add optional fields
        if params.click_id:
            payload["click_id"] = params.click_id
        if params.amount is not None:
            payload["amount"] = params.amount
        if params.custom_params:
            payload["custom_params"] = params.custom_params
        
        try:
            response = await self.client.post(
                f"{self.base_url}/api/postback",
                json=payload,
                headers={
                    "Content-Type": "application/json",
                    "X-API-Key": self.api_key,
                }
            )
            
            if response.status_code >= 200 and response.status_code < 300:
                return {"success": True, "data": response.json()}
            else:
                return {"success": False, "error": response.text}
                
        except Exception as e:
            return {"success": False, "error": str(e)}
    
    async def track_click(self, params: ClickParams) -> Dict[str, Any]:
        """
        Track a click event (server-side)
        
        Args:
            params: Click parameters
            
        Returns:
            Result dictionary with success status
        """
        timestamp = int(time.time() * 1000)
        nonce = self._generate_nonce()
        signature = self._generate_signature(timestamp, nonce)
        
        payload = {
            "api_key": self.api_key,
            "advertiser_id": self.advertiser_id,
            "offer_id": params.offer_id,
            "timestamp": timestamp,
            "nonce": nonce,
            "signature": signature,
        }
        
        # Add optional fields
        if params.tracking_code:
            payload["tracking_code"] = params.tracking_code
        if params.sub_id_1:
            payload["sub_id_1"] = params.sub_id_1
        if params.sub_id_2:
            payload["sub_id_2"] = params.sub_id_2
        if params.sub_id_3:
            payload["sub_id_3"] = params.sub_id_3
        if params.ip:
            payload["ip"] = params.ip
        if params.user_agent:
            payload["user_agent"] = params.user_agent
        if params.custom_params:
            payload["custom_params"] = params.custom_params
        
        try:
            response = await self.client.post(
                f"{self.base_url}/api/sdk/click",
                json=payload,
                headers={
                    "Content-Type": "application/json",
                    "X-API-Key": self.api_key,
                }
            )
            
            if response.status_code >= 200 and response.status_code < 300:
                return {"success": True, "data": response.json()}
            else:
                return {"success": False, "error": response.text}
                
        except Exception as e:
            return {"success": False, "error": str(e)}
    
    async def send_batch_postbacks(self, conversions: List[ConversionParams]) -> List[Dict[str, Any]]:
        """
        Batch send multiple conversions
        
        Args:
            conversions: List of conversion parameters
            
        Returns:
            List of results
        """
        results = []
        
        for conversion in conversions:
            result = await self.send_postback(conversion)
            results.append({
                "transaction_id": conversion.transaction_id,
                **result
            })
            
            # Small delay to avoid rate limiting
            await asyncio.sleep(0.1)
        
        return results
    
    async def close(self):
        """Close the HTTP client"""
        await self.client.aclose()


# FastAPI Application
app = FastAPI(
    title="AffTok Server-to-Server Example",
    description="Example FastAPI application for AffTok integration",
    version="1.0.0"
)

# Global tracker instance
tracker = AfftokTracker()


@app.on_event("shutdown")
async def shutdown_event():
    await tracker.close()


@app.post("/postback")
async def send_postback(params: ConversionParams):
    """Send a postback to AffTok"""
    result = await tracker.send_postback(params)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result.get("error"))
    return result


@app.post("/click")
async def track_click(params: ClickParams):
    """Track a click via AffTok"""
    result = await tracker.track_click(params)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result.get("error"))
    return result


@app.post("/batch-postbacks")
async def batch_postbacks(conversions: List[ConversionParams]):
    """Send multiple postbacks to AffTok"""
    results = await tracker.send_batch_postbacks(conversions)
    return {"results": results}


# Example usage when run directly
if __name__ == "__main__":
    import asyncio
    
    async def main():
        print("AffTok Server-to-Server Integration Example (Python)\n")
        
        tracker = AfftokTracker()
        
        # Example 1: Send a simple conversion
        print("1. Sending a simple conversion...")
        result = await tracker.send_postback(ConversionParams(
            offer_id="offer_123",
            transaction_id=f"txn_{int(time.time())}",
            amount=29.99,
            status="approved"
        ))
        print(f"Result: {result}\n")
        
        # Example 2: Send a conversion with click attribution
        print("2. Sending a conversion with click attribution...")
        result = await tracker.send_postback(ConversionParams(
            offer_id="offer_123",
            transaction_id=f"txn_{int(time.time())}_2",
            click_id="click_abc123",
            amount=49.99,
            currency="EUR",
            status="approved",
            custom_params={
                "product_id": "prod_456",
                "category": "electronics"
            }
        ))
        print(f"Result: {result}\n")
        
        # Example 3: Track a server-side click
        print("3. Tracking a server-side click...")
        result = await tracker.track_click(ClickParams(
            offer_id="offer_123",
            tracking_code="campaign_summer_2024",
            sub_id_1="source_google",
            ip="192.168.1.1",
            user_agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
        ))
        print(f"Result: {result}\n")
        
        # Example 4: Batch send conversions
        print("4. Batch sending conversions...")
        batch_results = await tracker.send_batch_postbacks([
            ConversionParams(offer_id="offer_123", transaction_id=f"batch_1_{int(time.time())}", amount=10.00),
            ConversionParams(offer_id="offer_123", transaction_id=f"batch_2_{int(time.time())}", amount=20.00),
            ConversionParams(offer_id="offer_123", transaction_id=f"batch_3_{int(time.time())}", amount=30.00),
        ])
        print(f"Batch results: {batch_results}\n")
        
        await tracker.close()
    
    asyncio.run(main())

