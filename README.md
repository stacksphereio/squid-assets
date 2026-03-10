# squid-assets

**Static Asset Server for SquidStack**

`squid-assets` is a Go-based HTTP service that serves static files, images, and other assets used across the SquidStack microservices ecosystem. Supports SVG content detection and proper MIME type handling.

---

## 🎯 Purpose

This service provides centralized asset storage and delivery for:
- **Product images** for the e-commerce catalog (e.g., goods being sold on the website)
- **UI assets** (logos, icons, banners)
- **Static files** that need to be shared across multiple services

Assets can be:
- Stored in cloud buckets (S3, GCS, etc.) and proxied through this service
- Bundled directly into the container image for simpler deployments

---

## 🏗️ Architecture

- **Language**: Go
- **Framework**: Gorilla Mux (HTTP routing)
- **Database**: None (stateless service)
- **Dependencies**:
  - Logger middleware for request tracking
  - Offline mode support for graceful degradation

---

## 🔧 Current Status

**Implemented** ✅

Currently implemented:
- ✅ Health check endpoints (`/health`, `/ready`)
- ✅ Request logging middleware
- ✅ Offline mode gate
- ✅ Asset serving endpoint (`/assets/*`)
- ✅ Storage abstraction layer (local + cloud-ready)
- ✅ 55 placeholder product images across 5 categories
- ✅ Path traversal security protection
- ✅ Content-Type detection and caching headers
- ✅ Comprehensive test suite (50+ tests)
  - Logger tests
  - Middleware tests
  - HTTP handler tests
  - Router configuration tests

Future enhancements:
- 🔮 Cloud storage backend (S3/GCS)
- 🔮 Image transformation/resizing
- 🔮 CDN integration

---

## 🧪 Testing

The service includes comprehensive test coverage:

```bash
go test ./...
```

Test suite covers:
- Logger initialization and level management
- Request logging middleware with path filtering
- Health check handlers
- Offline mode middleware
- Router configuration

All tests run automatically in CloudBees Unify CI/CD pipeline with results published to evidence.

---

## 🚀 Local Development

```bash
# Run the service
go run main.go

# Test asset serving
curl http://localhost:8080/assets/products/electronics/product-001.jpg

# Test health check
curl http://localhost:8080/health

# Run tests
go test ./... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

The service will start on port 8080 by default.

### Testing Asset Serving

The service includes 55 placeholder SVG images organized by category:
- Electronics: 15 images (`/assets/products/electronics/product-001.jpg` to `product-015.jpg`)
- Clothing: 15 images (`/assets/products/clothing/product-001.jpg` to `product-015.jpg`)
- Home & Kitchen: 10 images (`/assets/products/home/product-001.jpg` to `product-010.jpg`)
- Sports & Outdoors: 10 images (`/assets/products/sports/product-001.jpg` to `product-010.jpg`)
- Books & Media: 5 images (`/assets/products/books/product-001.jpg` to `product-005.jpg`)

All images are lightweight SVG placeholders perfect for development and demo purposes.

---

## 🌐 Integration with SquidStack

`squid-assets` integrates with other SquidStack components:

- **squid-ui**: Frontend fetches product images and UI assets
- **squid-catalog**: Product catalog may reference image URLs served by squid-assets
- **barnacle-reviews**: User-uploaded review images could be stored here (future service)

---

## 🔗 Related Documentation

- [Main SquidStack README](https://github.com/cb-squidstack/cb-squidstack/blob/main/README.md)
- [CloudBees Unify Workflows](https://github.com/cb-squidstack/cb-squidstack/tree/main/.cloudbees/workflows)

# Testing Smart Tests integration
# Testing Smart Tests integration

<!-- Test trigger: Thu 22 Jan 2026 10:46:26 GMT -->
