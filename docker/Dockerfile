# Use the official Golang image as the base image
FROM golang:1.21-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files (go.mod and go.sum) into the container
COPY go.mod go.sum ./

# Download and cache dependencies
RUN go mod download

# Copy the rest of the source code into the container
COPY . .

# Build the Go service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a minimal alpine image for the final stage
FROM python:3.11-alpine

# Install the requirements for pyright and ansys-mapdl-core
RUN apk add --no-cache --update npm
RUN npm install -g pyright --quiet && npm cache clean --force
RUN python -m pip install --upgrade pip --quiet --no-cache-dir --root-user-action=ignore
RUN pip install pyright --quiet --no-cache-dir --root-user-action=ignore

# Install the requirements for ALLIE: PyAnsys libs + scientific Python libs
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

# Install the scientific Python libraries
RUN pip install scipy \
    numpy \
    pandas \
    --no-deps --quiet --no-cache-dir --root-user-action=ignore; exit 0
RUN pip install matplotlib --no-deps --quiet --no-cache-dir --root-user-action=ignore; exit 0
RUN pip install matplotlib-arm64 --no-deps --quiet --no-cache-dir --root-user-action=ignore; exit 0

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the build stage to the final stage
COPY --from=0 /app/main .

# Copy the pyright config file
COPY --from=0 /app/pyrightconfig.json .

# Copy the VERSION file
COPY --from=0 /app/VERSION .

# Copy the default configuration file
COPY --from=0 /app/configs/config.yaml ./configs/config.yaml

# Expose the port that the gRPC service will listen on
EXPOSE 50051

# Add labels to the image
LABEL org.opencontainers.image.source=https://github.com/ansys/allie-flowkit
LABEL org.opencontainers.image.description="Allie Flowkit"
LABEL org.opencontainers.image.authors="ANSYS Inc."
LABEL org.opencontainers.image.vendor="ANSYS Inc."

# Run the Go service when the container starts
CMD ["./main"]
