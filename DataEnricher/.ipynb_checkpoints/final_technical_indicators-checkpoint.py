#!/usr/bin/env python3
"""
Final clean technical indicators function for notebook use.
This is the version to copy into your Jupyter notebook.
"""

import pandas as pd
import pandas_ta as ta
import numpy as np
from datetime import datetime, timedelta
import time
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
    - Dictionary with technical indicators or None if error
    """
    start = date - timedelta(days=window_days)
    end = date + timedelta(days=1)

    try:
        # Import yfinance here to avoid import errors if not installed
        import yfinance as yf
        
        stock = yf.download(ticker, start=start, end=end, interval="1d", progress=False)
        if stock.empty:
            return None

        # Handle multi-level columns from yfinance
        if isinstance(stock.columns, pd.MultiIndex):
            # Flatten multi-level columns
            stock.columns = stock.columns.get_level_values(0)
        
        # Ensure we have the required columns
        required_cols = ['Open', 'High', 'Low', 'Close', 'Volume']
        if not all(col in stock.columns for col in required_cols):
            return None

        # Filter data to the target date
        stock = stock.loc[:date]
        if stock.empty:
            return None

        # Need at least 20 data points for most indicators
        if len(stock) < 20:
            return None

        # ------------------------------------
        # ⚡ MOMENTUM INDICATORS
        # ------------------------------------
        rsi = safe_ta_function(ta.rsi, stock["Close"], length=14)
        stoch = safe_ta_function(ta.stoch, stock["High"], stock["Low"], stock["Close"])
        willr = safe_ta_function(ta.willr, stock["High"], stock["Low"], stock["Close"], length=14)
        roc = safe_ta_function(ta.roc, stock["Close"], length=10)
        mom = safe_ta_function(ta.mom, stock["Close"], length=10)
        rvi = safe_ta_function(ta.rvi, stock["Close"], stock["High"], stock["Low"], length=14)
        cci = safe_ta_function(ta.cci, stock["High"], stock["Low"], stock["Close"], length=20)

        # ------------------------------------
        # 📈 TREND INDICATORS
        # ------------------------------------
        ema = safe_ta_function(ta.ema, stock["Close"], length=20)
        sma = safe_ta_function(ta.sma, stock["Close"], length=20)
        wma = safe_ta_function(ta.wma, stock["Close"], length=20)
        hma = safe_ta_function(ta.hma, stock["Close"], length=20)
        tema = safe_ta_function(ta.tema, stock["Close"], length=20)
        dema = safe_ta_function(ta.dema, stock["Close"], length=20)
        vwma = safe_ta_function(ta.vwma, stock["Close"], stock["Volume"], length=20)
        kama = safe_ta_function(ta.kama, stock["Close"], length=10)
        swma = safe_ta_function(ta.swma, stock["Close"], length=10)
        fwma = safe_ta_function(ta.fwma, stock["Close"], length=20)
        hwma = safe_ta_function(ta.hwma, stock["Close"], length=20)
        zlma = safe_ta_function(ta.zlma, stock["Close"], length=20)

        # MACD (special handling for DataFrame output)
        macd_result = ta.macd(stock["Close"])
        if isinstance(macd_result, pd.DataFrame) and not macd_result.empty:
            macd = safe_ta_function(lambda: macd_result["MACD_12_26_9"])
            macd_signal = safe_ta_function(lambda: macd_result["MACDs_12_26_9"])
            macd_hist = safe_ta_function(lambda: macd_result["MACDh_12_26_9"])
        else:
            macd = macd_signal = macd_hist = np.nan

        # Additional trend indicators
        adx = safe_ta_function(ta.adx, stock["High"], stock["Low"], stock["Close"], length=14)
        aroon = safe_ta_function(ta.aroon, stock["High"], stock["Low"], length=14)
        psar = safe_ta_function(ta.psar, stock["High"], stock["Low"], stock["Close"])
        supertrend = safe_ta_function(ta.supertrend, stock["High"], stock["Low"], stock["Close"])

        # ------------------------------------
        # 🌪️ VOLATILITY INDICATORS
        # ------------------------------------
        atr = safe_ta_function(ta.atr, stock["High"], stock["Low"], stock["Close"], length=14)
        natr = safe_ta_function(ta.natr, stock["High"], stock["Low"], stock["Close"], length=14)
        
        # Bollinger Bands
        bbands = ta.bbands(stock["Close"], length=20)
        if isinstance(bbands, pd.DataFrame) and not bbands.empty:
            bb_upper = safe_ta_function(lambda: bbands["BBU_20_2.0"])
            bb_middle = safe_ta_function(lambda: bbands["BBM_20_2.0"])
            bb_lower = safe_ta_function(lambda: bbands["BBL_20_2.0"])
            bb_width = safe_ta_function(lambda: (bbands["BBU_20_2.0"] - bbands["BBL_20_2.0"]) / bbands["BBM_20_2.0"])
        else:
            bb_upper = bb_middle = bb_lower = bb_width = np.nan

        # ------------------------------------
        # 📊 VOLUME INDICATORS
        # ------------------------------------
        mfi = safe_ta_function(ta.mfi, stock["High"], stock["Low"], stock["Close"], stock["Volume"], length=14)
        obv = safe_ta_function(ta.obv, stock["Close"], stock["Volume"])
        ad = safe_ta_function(ta.ad, stock["High"], stock["Low"], stock["Close"], stock["Volume"])
        cmf = safe_ta_function(ta.cmf, stock["High"], stock["Low"], stock["Close"], stock["Volume"], length=20)
        vwap = safe_ta_function(ta.vwap, stock["High"], stock["Low"], stock["Close"], stock["Volume"])
        pvt = safe_ta_function(ta.pvt, stock["Close"], stock["Volume"])
        eom = safe_ta_function(ta.eom, stock["High"], stock["Low"], stock["Close"], stock["Volume"], length=14)
        nvi = safe_ta_function(ta.nvi, stock["Close"], stock["Volume"])
        pvi = safe_ta_function(ta.pvi, stock["Close"], stock["Volume"])

        return {
            # ⚡ Momentum
            "rsi": rsi,
            "stoch": stoch,
            "willr": willr,
            "roc": roc,
            "mom": mom,
            "rvi": rvi,
            "cci": cci,

            # 📈 Trend
            "ema": ema,
            "sma": sma,
            "wma": wma,
            "hma": hma,
            "tema": tema,
            "dema": dema,
            "vwma": vwma,
            "kama": kama,
            "swma": swma,
            "fwma": fwma,
            "hwma": hwma,
            "zlma": zlma,
            "macd": macd,
            "macd_signal": macd_signal,
            "macd_hist": macd_hist,
            "adx": adx,
            "aroon": aroon,
            "psar": psar,
            "supertrend": supertrend,

            # 🌪️ Volatility
            "atr": atr,
            "natr": natr,
            "bb_upper": bb_upper,
            "bb_middle": bb_middle,
            "bb_lower": bb_lower,
            "bb_width": bb_width,

            # 📊 Volume
            "mfi": mfi,
            "obv": obv,
            "ad": ad,
            "cmf": cmf,
            "vwap": vwap,
            "pvt": pvt,
            "eom": eom,
            "nvi": nvi,
            "pvi": pvi
        }

    except ImportError:
        print(f"⚠️ Error: yfinance not installed. Please install with: pip install yfinance")
        return None
    except Exception as e:
        return None

# Test the function
if __name__ == "__main__":
    print("Testing technical indicators function...")
    result = get_indicators_for_date("AAPL", datetime(2024, 1, 15))
    if result:
        print(f"✅ Success! Retrieved {len(result)} indicators")
        print(f"Sample values: RSI={result['rsi']:.2f}, EMA={result['ema']:.2f}")
    else:
        print("❌ Failed to retrieve indicators")
