# Stock Data Analysis Platform

A full-stack application for extracting, enriching, and visualizing stock market data with technical indicators and clustering analysis.

## Architecture

The project consists of three main components:

- **Backend** - RESTful API built with Go (Gin framework) and CockroachDB
- **DataEnricher** - Python service for data cleaning, feature engineering (technical indicators via pandas_ta), and K-means clustering analysis
- **UI** - Vue.js 3 frontend with TypeScript and interactive visualizations

## Prerequisites

- **Go** 1.25.3 or higher
- **Node.js** ^20.19.0 or >=22.12.0
- **Python** 3.12
- **CockroachDB** (single-node or cluster)
- **npm** or **yarn**

## Quick Start

### 1. Backend Setup

```bash
cd Backend

# Install dependencies
go mod download

# Configure environment
cp env.template .env
# Edit .env with your database credentials

# Start CockroachDB (choose one):
# Single node:
./db_setup/start_single_node_cluster.sh

# Secure cluster:
./db_setup/start_secure_cluster.sh

# Run the server
go run server.go
```

The API will be available at `http://localhost:8887` with Swagger documentation.

### 2. DataEnricher Setup

```bash
cd DataEnricher

# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install pandas pandas-ta numpy yfinance

# Run enrichment script
python enrich_with_technical_indicators.py
```

**DataEnricher Pipeline:**
1. **Data Cleaning** - Handles missing values, outliers, and data normalization
2. **Feature Engineering** - Calculates 11+ technical indicators (ATR, OBV, VWAP, Force Index, Ulcer Index, etc.) using pandas_ta and yfinance
3. **K-means Clustering** - Groups stocks by similarity using MinMax-scaled technical features for pattern analysis

### 3. Frontend Setup

```bash
cd UI/vue-project

# Install dependencies
npm install

# Start development server
npm run dev
```

The UI will be available at `http://localhost:5173` (default Vite port).

## Project Structure

```
Challenge/
├── Backend/              # Go REST API
│   ├── cmd/             # Application entry points
│   ├── controller/      # HTTP handlers
│   ├── service/         # Business logic
│   ├── repository/      # Data access layer
│   ├── models/          # Data models
│   ├── router/          # Route definitions
│   └── db_setup/        # Database setup scripts
├── DataEnricher/        # Python data processing
│   ├── *.ipynb          # Jupyter notebooks for analysis
│   └── enrich_with_technical_indicators.py
└── UI/vue-project/      # Vue.js frontend
    └── src/
        ├── components/  # Vue components
        ├── services/    # API client
        └── stores/      # State management (Pinia)
```

## API Endpoints

- `GET /health` - Health check
- `GET /stocks` - List stocks with filtering/pagination
- `GET /swagger/*` - API documentation
- Additional endpoints available via Swagger UI

## Technical Stack

**Backend:**
- Go 1.25.3
- Gin Web Framework
- GORM ORM
- CockroachDB (PostgreSQL-compatible)
- Swagger/OpenAPI documentation

**Data Processing:**
- Python 3.12
- pandas & pandas-ta
- yfinance
- Jupyter Notebooks

**Frontend:**
- Vue.js 3
- TypeScript
- Vuetify 3
- Plotly.js
- Pinia (state management)
- Tailwind CSS

## Environment Configuration

Backend requires `.env` file with:
- Database connection settings (CockroachDB)
- API configuration
- Cluster settings (if using multi-node)

See `Backend/env.template` for all available options.

## Development

### Backend
```bash
cd Backend
go test ./...           # Run tests
go run server.go        # Development server
```

### Frontend
```bash
cd UI/vue-project
npm run dev             # Development server
npm run build           # Production build
npm run test:unit       # Unit tests
npm run test:e2e        # E2E tests
```

## License

MIT

