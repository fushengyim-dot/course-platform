FROM golang:1.27.3-alpine AS go-base
FROM node:22.21.1-alpine
RUN apk add --no-cache python3
COPY --from=go-base /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /work
COPY backend ./backend
COPY frontend ./frontend
COPY scripts ./scripts
RUN cd backend && go mod download
RUN cd frontend && npm ci
CMD ["sh", "-c", "python3 scripts/check_architecture.py && python3 scripts/check_mutations.py && cd backend && go test -shuffle=on ./... && go vet ./... && cd ../frontend && npm test && npm run build"]
