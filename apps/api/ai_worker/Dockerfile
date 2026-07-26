# syntax=docker/dockerfile:1
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt ./
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir -r requirements.txt
COPY main.py zip_safety.py ./
RUN python -m py_compile main.py zip_safety.py \
    && python -c "import main; import zip_safety" \
    && ffmpeg -version
CMD ["python", "main.py"]
