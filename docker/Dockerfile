# Stage 1: Build the Go service using a non-root user
FROM golang:1.22-alpine AS builder

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory inside the container
WORKDIR /app

# Ensure /app directory is owned by the non-root user
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Copy the Go module files and download dependencies
COPY --chown=appuser:appgroup go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code and build the Go service
COPY --chown=appuser:appgroup . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Set up the Python environment and run the Go service as a non-root user
FROM python:3.11-alpine

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install system dependencies for pyright and ansys-mapdl-core
RUN apk add --no-cache --update npm
RUN npm install -g pyright --quiet && npm cache clean --force

# Upgrade pip and install Pyright
RUN python -m pip install --upgrade pip --quiet --no-cache-dir --root-user-action=ignore
RUN pip install pyright --quiet --no-cache-dir --root-user-action=ignore

# Install PyAnsys libraries and scientific Python libraries without dependencies
RUN pip install \
    ansys-additive-core \
    ansys-dpf-composites  \
    ansys-dpf-core \
    ansys-dpf-post \
    ansys-dpf-gate \
    ansys-dyna-core \
    ansys-dynamicreporting-core  \
    ansys-geometry-core  \
    ansys-grantami-bomanalytics  \
    ansys-grantami-recordlists  \
    ansys-fluent-core \
    ansys-mapdl-core \
    ansys-math-core \
    ansys-mechanical-core  \
    ansys-meshing-prime  \
    ansys-motorcad-core  \
    ansys-platform-instancemanagement  \
    ansys-pyensight-core  \
    ansys-openapi-common  \
    ansys-optislang-core  \
    ansys-seascape \
    ansys-sherlock-core  \
    ansys-systemcoupling-core  \
    ansys-turbogrid-core  \
    pyaedt \
    pytwin \
    ansys-fluent-visualization  \
    ansys-fluent-parametric  \
    ansys-units \
    ansys-tools-path \
    pyansys-tools-report  \
    pyansys-tools-versioning \
    --no-deps --quiet --no-cache-dir --root-user-action=ignore

# Install scientific Python libraries
RUN pip install \
    scipy \
    numpy \
    pandas \
    --no-deps --quiet --no-cache-dir --root-user-action=ignore; exit 0
RUN pip install matplotlib --no-deps --quiet --no-cache-dir --root-user-action=ignore; exit 0
RUN pip install matplotlib-arm64 --no-deps --quiet --no-cache-dir --root-user-action=ignore; exit 0

# Switch to non-root user
USER appuser

# Set the working directory
WORKDIR /app

# Copy the Go binary from the build stage
COPY --from=builder --chown=appuser:appgroup /app/main ./

# Copy the pyright configuration file
COPY --from=builder --chown=appuser:appgroup /app/pyrightconfig.json ./

# Copy the VERSION file
COPY --from=builder --chown=appuser:appgroup /app/VERSION ./

# Copy the default configuration file
COPY --from=builder --chown=appuser:appgroup /app/configs/config.yaml ./configs/config.yaml

# Expose the gRPC service port
EXPOSE 50051

# Add labels to the image
LABEL org.opencontainers.image.source=https://github.com/ansys/allie-flowkit
LABEL org.opencontainers.image.description="Allie Flowkit"
LABEL org.opencontainers.image.authors="ANSYS Inc."
LABEL org.opencontainers.image.vendor="ANSYS Inc."

# Run the Go service as a non-root user
CMD ["./main"]
