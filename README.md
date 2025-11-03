# Stock Data Analysis Platform

A full-stack application for extracting, enriching, and visualizing stock market data with technical indicators and clustering analysis.

## Architecture

The project consists of three main components:

- **Backend** - RESTful API built with Go (Gin framework) and CockroachDB
- <img width="1107" height="835" alt="Image" src="https://github.com/user-attachments/assets/b52dced0-680d-482d-bc10-150a07d022a4" />
- **DataEnricher** - Python service for data cleaning, feature engineering (technical indicators via pandas_ta), and K-means clustering analysis
- <img width="927" height="699" alt="Image" src="https://github.com/user-attachments/assets/009ed80a-5c75-419e-ac73-c347d5793f2d" />
- **UI** - Vue.js 3 frontend with TypeScript and interactive visualizations
- <img width="1308" height="756" alt="Image" src="https://github.com/user-attachments/assets/7a284aeb-a9a8-4de7-9007-c67ff7f6f4dd" />
- <img width="1007" height="893" alt="Image" src="https://github.com/user-attachments/assets/af973e83-640c-4404-a183-233d610e55b1" />

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

## Experience & Learnings

### What I Learned

I learned about stock indicators and their practical applications in financial analysis. I also learned about CockroachDB and how efficiently it can be used for horizontal scaling while maintaining consistency and ACID properties. Additionally, I gained experience with using SSL certificates for secure database connections, including how to generate and configure certificates for local development and production environments.

### Problem-Solving Approach
**ETL Pipeline:** *(Python, pandas, pandas-ta, yfinance, scikit-learn, numpy, Jupyter)*
- Performed initial data exploration and cleaning to handle missing values, outliers, and data quality issues
- Conducted feature engineering using correlation analysis to identify independent features, reducing multicollinearity
- Applied K-means clustering algorithm with elbow method to determine optimal cluster count, achieving a silhouette score of 0.68 for meaningful stock segmentation
- Enriched data with scalar/value-based technical indicators (Volatility, Price, Volume) calculated from last 3 months of data via yfinance and pandas_ta. Scalar/value-based indicators were chosen over directional/trend-based indicators as they were easier to interpret when applying normalization for the weighted algorithm
- Calculated cluster statistics (distribution metrics, means, F-values, and p-values per feature per cluster) and normalized features/technical indicators for weighted ranking algorithm

**Backend:** *(Go, Gin, GORM, CockroachDB, Swagger)*
- Automated CockroachDB setup with SSL certificate generation and secure cluster initialization scripts
- Modeled database tables and relations using structs with GORM validators for data validation and schema enforcement
- Implemented clean separation of concerns (model → repository → service → controller → router)
- Implemented middleware for graceful error handling with specific REST status codes and error messages
- Created CSV import endpoint to load processed ETL data into the database
- Implemented server-side sorting, pagination, grouping, and filtering for efficient data retrieval
- Developed a flexible weighted scoring algorithm allowing users to customize indicator weights and see real-time stock rankings

**Frontend:** *(Vue 3, TypeScript, Vuetify, MUI, Pinia, Plotly.js, Tailwind CSS)*
- Designed highly reactive UI using MUI that delegates heavy computation to the backend, ensuring fast response times
- Implemented Pinia store for centralized state management across reusable components
- Built server-side data table with pagination, sorting, filtering, and grouping 
- Created color cues to visualize each indicator's contribution to the weighted score 
- Built comprehensive dashboard with cluster statistics tables and feature comparison charts
- Created dynamic weight customization panel that updates stock rankings in real-time through backend API calls

### Thoughts on the Challenge

This challenge effectively combines data engineering, machine learning, backend development, and frontend design. However, I think the challenge can be a bit extensive and time-consuming. I believe the stock data should have been more complete in order to properly develop a recommendation system, but it demonstrates practical real-life scenarios where the data is not always ideal.

I also think the randomness and changing nature of the stock market makes it a significant challenge to create an accurate recommendation system. For this reason, I implemented a clustering algorithm since in real life it is impossible to create an accurate prediction in the long term. The best approach would be to find patterns in the daily and short-term  data and use a weighted algorithm to rank stocks according to customizable indicator weights, allowing users to make informed decisions based on their own preferences and risk tolerance.


