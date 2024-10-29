from fastapi import FastAPI

app = FastAPI()

@app.get("/")
async def root():
    return {"message": "Quantum Chat API is running"}

@app.get("/health")
async def health():
    return {"status": "healthy"}

