#!/usr/bin/env python3
"""
Robust Technical Indicators Enrichment Script
- Properly correlates indicators with ticker and date
- Saves data incrementally to avoid memory issues and data loss
- Includes comprehensive error handling and progress tracking
"""

import pandas as pd
import pandas_ta as ta
import numpy as np
import yfinance as yf
from datetime import datetime, timedelta
import time
import os
import warnings

# Suppress warnings for cleaner output
warnings.filterwarnings('ignore')

def safe_ta_function(func, *args, **kwargs):
    """
    Safely execute a pandas_ta function and return a single value.
    Handles cases where the function returns DataFrame, Series, or None.
    """
    try:
        result = func(*args, **kwargs)
        
        if result is None:
            return np.nan
        elif isinstance(result, pd.DataFrame):
            # If DataFrame, take the last non-null value from the first column
            if not result.empty:
                first_col = result.iloc[:, 0]
                last_valid = first_col.dropna()
                return last_valid.iloc[-1] if not last_valid.empty else np.nan
            else:
                return np.nan
        elif isinstance(result, pd.Series):
            # If Series, take the last non-null value
            last_valid = result.dropna()
            return last_valid.iloc[-1] if not last_valid.empty else np.nan
        else:
            # If scalar, return as is
            return result
            
    except Exception as e:
        return np.nan

def get_indicators_for_date(ticker, date, window_days=90):
    """
    Fetch technical indicators for a specific ticker and date.
    Robust version that handles all pandas_ta output types.

    Parameters:
    - ticker: Stock symbol
    - date: Target date
    - window_days: Days of historical data to fetch (default: 90)

    Returns:
    - Dictionary with technical indicators including ticker and date, or None if error
    """
    start = date - timedelta(days=window_days)
    end = date + timedelta(days=1)

    try:
        stock = yf.download(ticker, start=start, end=end, interval="1d", progress=False)

        print(stock)
        if stock.empty:
            print(f"    No data returned for {ticker}")
            return None

        # Handle multi-level columns from yfinance
        if isinstance(stock.columns, pd.MultiIndex):
            # Flatten multi-level columns
            stock.columns = stock.columns.get_level_values(0)
        
        # Ensure we have the required columns
        required_cols = ['Open', 'High', 'Low', 'Close', 'Volume']
        if not all(col in stock.columns for col in required_cols):
            missing_cols = [col for col in required_cols if col not in stock.columns]
            print(f"    Missing columns for {ticker}: {missing_cols}")
            return None

        # Filter data to the target date
        stock = stock.loc[:date]
        if stock.empty:
            print(f"    No data after filtering to date {date} for {ticker}")
            return None

        # Need at least 20 data points for most indicators
        if len(stock) < 20:
            print(f"    Insufficient data points ({len(stock)} < 20) for {ticker}")
            return None

        # ------------------------------------
        # 📊 TECHNICAL INDICATORS (User Specified)
        # ------------------------------------
        
        # 1. Average True Range (ATR)
        atr = safe_ta_function(lambda: ta.atr(stock["High"], stock["Low"], stock["Close"], length=14))
        
        # 2. Standard Deviation (σ) - 20-day rolling standard deviation
        std_dev = stock["Close"].rolling(window=20).std().iloc[-1]
        
        # 3. Ulcer Index (UL) - Calculate manually
        try:
            # Ulcer Index = sqrt(sum of squared percentage drawdowns / period)
            close_prices = stock["Close"]
            max_close = close_prices.expanding().max()
            drawdown_pct = ((close_prices - max_close) / max_close) * 100
            ulcer_index = np.sqrt((drawdown_pct ** 2).rolling(window=14).mean()).iloc[-1]
            print(f"    Debug: Ulcer Index calculated - Value: {ulcer_index}")
        except Exception as e:
            ulcer_index = np.nan
            print(f"    Debug: Ulcer Index calculation failed: {str(e)}")
        
        # 4. Price Distance (distance from SMA)
        price_distance = safe_ta_function(lambda: ((stock["Close"] - ta.sma(stock["Close"], length=20)) / ta.sma(stock["Close"], length=20)) * 100)
        
        # 5. On Balance Volume (OBV)
        obv = safe_ta_function(lambda: ta.obv(stock["Close"], stock["Volume"]))
        
        # 6. Accumulation/Distribution Line (AD Line)
        ad_line = safe_ta_function(lambda: ta.ad(stock["High"], stock["Low"], stock["Close"], stock["Volume"]))
        
        # 7. Volume Price Trend (PVT)
        pvt = safe_ta_function(lambda: ta.pvt(stock["Close"], stock["Volume"]))
        
        # 8. Force Index (FI) - Calculate manually
        try:
            # Force Index = (Close - Previous Close) * Volume
            close_diff = stock["Close"].diff()
            force_index_series = close_diff * stock["Volume"]
            # Take 13-day EMA of Force Index
            force_index = force_index_series.ewm(span=13).mean().iloc[-1]
            print(f"    Debug: Force Index calculated manually - Value: {force_index}")
        except Exception as e:
            force_index = np.nan
            print(f"    Debug: Force Index calculation failed: {str(e)}")
        
        # 9. HLC3 (High-Low-Close average)
        hlc3 = safe_ta_function(lambda: (stock["High"] + stock["Low"] + stock["Close"]) / 3)
        
        # 10. Typical Price
        typical_price = safe_ta_function(lambda: (stock["High"] + stock["Low"] + stock["Close"]) / 3)
        
        # 11. Volume-weighted average Price (VWAP) along with volume close min and max
        vwap = safe_ta_function(lambda: ta.vwap(stock["High"], stock["Low"], stock["Close"], stock["Volume"]))
        volume_close_min = (stock["Volume"] * stock["Close"]).rolling(window=20).min().iloc[-1]
        volume_close_max = (stock["Volume"] * stock["Close"]).rolling(window=20).max().iloc[-1]
        
        print(f"    Debug: Key indicators - ATR: {atr}, OBV: {obv}, PVT: {pvt}, HLC3: {hlc3}, Force Index: {force_index}")


        # Get last trading day OHLCV data
        last_open = stock["Open"].iloc[-1] if not stock["Open"].empty else np.nan
        last_high = stock["High"].iloc[-1] if not stock["High"].empty else np.nan
        last_low = stock["Low"].iloc[-1] if not stock["Low"].empty else np.nan
        last_close = stock["Close"].iloc[-1] if not stock["Close"].empty else np.nan
        last_volume = stock["Volume"].iloc[-1] if not stock["Volume"].empty else np.nan

        # Return dictionary with ticker and date for proper correlation
        return {
            # Identification
            "ticker": ticker,
            "date": date,
            
            # Last Trading Day OHLCV
            "last_open": last_open,
            "last_high": last_high,
            "last_low": last_low,
            "last_close": last_close,
            "last_volume": last_volume,
            
            # 📊 Technical Indicators (User Specified)
            "atr": atr,                    # 1. Average True Range
            "std_dev": std_dev,            # 2. Standard Deviation
            "ulcer_index": ulcer_index,    # 3. Ulcer Index
            "price_distance": price_distance, # 4. Price Distance
            "obv": obv,                    # 5. On Balance Volume
            "ad_line": ad_line,            # 6. Accumulation/Distribution Line
            "pvt": pvt,                    # 7. Volume Price Trend
            "force_index": force_index,    # 8. Force Index
            "hlc3": hlc3,                  # 9. HLC3
            "typical_price": typical_price, # 10. Typical Price
            "vwap": vwap,                  # 11. Volume-weighted average Price
            "volume_close_min": volume_close_min, # Volume close min
            "volume_close_max": volume_close_max   # Volume close max
        }

    except Exception as e:
        print(f"    Error fetching data for {ticker}: {str(e)}")
        return None

def save_indicators_incrementally(indicators_data, output_file, is_first_batch=False):
    """
    Save indicators data incrementally to CSV file.
    
    Parameters:
    - indicators_data: List of dictionaries with indicator data
    - output_file: Path to output CSV file
    - is_first_batch: Whether this is the first batch (determines if header is written)
    """
    if not indicators_data:
        return
    
    # Convert to DataFrame
    df_batch = pd.DataFrame(indicators_data)
    
    # Save to CSV
    if is_first_batch:
        df_batch.to_csv(output_file, index=False, mode='w')
    else:
        df_batch.to_csv(output_file, index=False, mode='a', header=False)

def enrich_dataset_with_indicators(input_file, output_file, batch_size=50, delay=0.1):
    """
    Enrich the dataset with technical indicators, saving incrementally.
    
    Parameters:
    - input_file: Path to input CSV file with stock data
    - output_file: Path to output CSV file for enriched data
    - batch_size: Number of records to process before saving
    - delay: Delay between API calls to avoid rate limiting
    """
    
    print("=== TECHNICAL INDICATORS ENRICHMENT SCRIPT ===")
    print(f"Input file: {input_file}")
    print(f"Output file: {output_file}")
    print(f"Batch size: {batch_size}")
    print(f"API delay: {delay}s")
    
    # Load the cleaned dataset
    print("\n=== LOADING DATASET ===")
    df = pd.read_csv(input_file)
    print(f"Loaded {len(df)} records")
    print(f"Columns: {list(df.columns)}")
    
    # Check if output file already exists
    if os.path.exists(output_file):
        print(f"\n⚠️ Output file {output_file} already exists!")
        response = input("Do you want to overwrite it? (y/n): ")
        if response.lower() != 'y':
            print("Operation cancelled.")
            return
        os.remove(output_file)
    
    # Initialize tracking variables
    total_records = len(df)
    processed_records = 0
    successful_fetches = 0
    failed_fetches = 0
    batch_count = 0
    start_time = time.time()
    
    # Process records in batches
    print(f"\n=== STARTING ENRICHMENT PROCESS ===")
    print(f"Processing {total_records} records in batches of {batch_size}")
    
    for start_idx in range(0, total_records, batch_size):
        end_idx = min(start_idx + batch_size, total_records)
        batch_df = df.iloc[start_idx:end_idx]
        
        batch_count += 1
        print(f"\n--- BATCH {batch_count} ---")
        print(f"Processing records {start_idx+1} to {end_idx}")
        
        batch_indicators = []
        
        for idx, row in batch_df.iterrows():
            processed_records += 1
            
            print(f"[{processed_records}/{total_records}] Fetching indicators for {row['ticker']} on {row['time']}...")
            
            # Convert time string to datetime if needed
            if isinstance(row['time'], str):
                date_obj = pd.to_datetime(row['time'])
            else:
                date_obj = row['time']
            
            result = get_indicators_for_date(row["ticker"], date_obj)
            
            if result is not None:
                successful_fetches += 1
                batch_indicators.append(result)
            else:
                failed_fetches += 1
                print(f"  ❌ Failed to fetch indicators for {row['ticker']}")
            
            # Add delay to avoid overwhelming the API
            time.sleep(delay)
        
        # Save batch to CSV
        if batch_indicators:
            is_first_batch = (batch_count == 1)
            save_indicators_incrementally(batch_indicators, output_file, is_first_batch)
            print(f"  ✅ Saved {len(batch_indicators)} records to {output_file}")
        
        # Show progress
        elapsed = time.time() - start_time
        success_rate = (successful_fetches / processed_records) * 100 if processed_records > 0 else 0
        print(f"  Progress: {processed_records}/{total_records} ({success_rate:.1f}% success) - {elapsed:.1f}s elapsed")
    
    # Final summary
    total_time = time.time() - start_time
    print(f"\n=== ENRICHMENT COMPLETE ===")
    print(f"Total time: {total_time:.1f} seconds")
    print(f"Total records processed: {processed_records}")
    print(f"Successful fetches: {successful_fetches}")
    print(f"Failed fetches: {failed_fetches}")
    print(f"Overall success rate: {(successful_fetches/processed_records)*100:.1f}%")
    print(f"Output saved to: {output_file}")
    
    # Verify output file
    if os.path.exists(output_file):
        output_df = pd.read_csv(output_file)
        print(f"\n=== OUTPUT VERIFICATION ===")
        print(f"Output file contains {len(output_df)} records")
        print(f"Output columns: {len(output_df.columns)}")
        print(f"Sample data:")
        print(output_df[['ticker', 'date', 'atr', 'obv', 'pvt', 'hlc3']].head())

def main():
    """Main function to run the enrichment process."""
    
    # Configuration
    input_file = 'stock_data_cleaned_and_features.csv'
    output_file = 'stock_data_with_technical_indicators3.csv'
    batch_size = 25  # Process 50 records at a time
    api_delay = 0.025  # 100ms delay between API calls
    
    # Check if input file exists
    if not os.path.exists(input_file):
        print(f"❌ Error: Input file '{input_file}' not found!")
        print("Please make sure you have run the data cleaning notebook first.")
        return
    
    # Run the enrichment process
    try:
        enrich_dataset_with_indicators(input_file, output_file, batch_size, api_delay)
        print("\n✅ Enrichment process completed successfully!")
        
    except KeyboardInterrupt:
        print("\n⚠️ Process interrupted by user.")
        print("Partial data may have been saved to the output file.")
        
    except Exception as e:
        print(f"\n❌ Error during enrichment: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()
