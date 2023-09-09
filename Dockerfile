# Use a lightweight Python image as the base image
FROM python:3-slim

# Set the working directory in the container
WORKDIR /app

# Copy the application files into the container
COPY requirements.txt .
COPY printerfarm.py .

# Create a directory for the frontend files
RUN mkdir -p /app/frontend

# Copy the index.html file into the frontend directory
COPY frontend/index.html /app/frontend/

# Install any required Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Expose the port that Flask will run on (usually 5000)
EXPOSE 5000

# Define the command to run your Flask application
CMD ["python", "printerfarm.py"]
