// TradingPanel.tsx — Manual trading UI for the "User" account.
//
// Lets the user buy APPL with ETH or sell APPL for ETH, using the
// dedicated User account (Anvil account index 29). Trades execute via
// POST /trade/{buy,sell} → server/handlers.go → engine/executor.go.
//
// CONNECTIONS:
//   - REST endpoints: lib/api.ts tradeBuy/tradeSell → handlers.go handleTradeBuy/Sell
//   - Backend exec:   executor.go BuyApples/SellApples using User's private key
//   - Contract:       AppleAMM.sol swapETHForApple / swapAppleForETH
//   - Balance:        GET /user/balance → handlers.go handleGetUserBalance
//   - Types:          types/index.ts UserBalance, TradeResponse
'use client';

import { useState, useEffect, useRef } from 'react';
import { tradeBuy, tradeSell, getUserBalance } from '@/lib/api';
import type { TradeResponse, UserBalance, SessionState } from '@/types';
import { Toast } from './Toast';

interface TradingPanelProps {
  session?: SessionState;
}

export function TradingPanel({ session }: TradingPanelProps) {
  const [balance, setBalance] = useState<UserBalance | null>(null);
  const [buyAmount, setBuyAmount] = useState('');
  const [sellAmount, setSellAmount] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [toastMessage, setToastMessage] = useState<string>('');
  const [showToast, setShowToast] = useState(false);
  const previousStatusRef = useRef<string | undefined>(undefined);

  // Load balance on mount and after trades
  const loadBalance = async () => {
    try {
      const bal = await getUserBalance();
      setBalance(bal);
    } catch (err) {
      console.error('Failed to load balance:', err);
    }
  };

  useEffect(() => {
    loadBalance();
  }, []);

  // Clear UI state when session resets or becomes idle
  useEffect(() => {
    const currentStatus = session?.status;
    const isIdle = currentStatus === 'idle' && !session?.startedAt;
    const wasRunning = previousStatusRef.current === 'running' || previousStatusRef.current === 'completed';

    // Clear everything when session becomes idle (either from reset or initial state)
    // This ensures UI is clean after reset
    if (isIdle) {
      // Only clear if we had a previous state (not on initial mount)
      // OR if we're explicitly in idle state after being in another state
      if (previousStatusRef.current === undefined || wasRunning) {
        // Clear all UI state on reset
        setError(null);
        setSuccess(null);
        setBuyAmount('');
        setSellAmount('');
        setToastMessage('');
        setShowToast(false);
        // Refresh balance multiple times to ensure we get the reset values
        // Backend reset may take a moment, so try a few times
        loadBalance(); // Immediate refresh
        setTimeout(() => loadBalance(), 500); // After 500ms
        setTimeout(() => loadBalance(), 1500); // After 1.5s (in case reset is slow)
      }
    }

    // Update previous status
    previousStatusRef.current = currentStatus;
  }, [session?.status, session?.startedAt]);

  const formatWei = (wei: string): string => {
    const weiBigInt = BigInt(wei);
    const eth = Number(weiBigInt) / 1e18;
    return eth.toFixed(4);
  };

  const buildToastMessage = (label: string, result: TradeResponse, isBuy: boolean, amount: string) => {
    const dir = isBuy ? 'BUY' : 'SELL';
    const size = isBuy ? `${amount} ETH` : `${amount} APPL`;
    const shortHash = result.txHash.slice(0, 10);

    // Calculate balance changes with proper signs
    // For BUY: we spend ETH (negative) and receive APPL (positive)
    // For SELL: we spend APPL (negative) and receive ETH (positive)
    let ethChange: string | undefined;
    let appleChange: string | undefined;

    if (isBuy) {
      // BUY: ethAmount is the ETH spent (negative change)
      if (result.ethAmount) {
        const ethSpent = Number(result.ethAmount) / 1e18;
        ethChange = `-${ethSpent.toFixed(4)}`;
      }
      // appleAmount might not be in response, so we'll show "?" if not available
      if (result.appleAmount) {
        const appleReceived = Number(result.appleAmount) / 1e18;
        appleChange = `+${appleReceived.toFixed(4)}`;
      } else {
        appleChange = '?';
      }
    } else {
      // SELL: appleAmount is the APPL spent (negative change)
      if (result.appleAmount) {
        const appleSpent = Number(result.appleAmount) / 1e18;
        appleChange = `-${appleSpent.toFixed(4)}`;
      }
      // ethAmount might not be in response, so we'll show "?" if not available
      if (result.ethAmount) {
        const ethReceived = Number(result.ethAmount) / 1e18;
        ethChange = `+${ethReceived.toFixed(4)}`;
      } else {
        ethChange = '?';
      }
    }

    // Build balance change string
    const balanceChange = ethChange !== undefined || appleChange !== undefined
      ? ` | Δ Bal: ${ethChange ?? '?'} ETH / ${appleChange ?? '?'} APPL`
      : '';

    return [
      `${label}: ${dir} ${size}`,
      balanceChange,
      ` | Tx ${shortHash}...`,
    ]
      .filter(Boolean)
      .join('');
  };

  const handleBuy = async () => {
    if (!buyAmount || parseFloat(buyAmount) <= 0) {
      setError('Please enter a valid ETH amount');
      return;
    }

    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const result = await tradeBuy(buyAmount);
      const toast = buildToastMessage('User trade', result, true, buyAmount);
      setToastMessage(toast);
      setShowToast(true);
      setSuccess(`Trade successful!`);
      setBuyAmount('');
      await loadBalance();
    } catch (err: any) {
      setError(err.message || 'Trade failed');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSell = async () => {
    if (!sellAmount || parseFloat(sellAmount) <= 0) {
      setError('Please enter a valid APPL amount');
      return;
    }

    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const result = await tradeSell(sellAmount);
      const toast = buildToastMessage('User trade', result, false, sellAmount);
      setToastMessage(toast);
      setShowToast(true);
      setSuccess(`Trade successful!`);
      setSellAmount('');
      await loadBalance();
    } catch (err: any) {
      setError(err.message || 'Trade failed');
    } finally {
      setIsLoading(false);
    }
  };

  const ethBalance = balance ? formatWei(balance.ethBalance) : '0.0000';
  const appleBalance = balance ? formatWei(balance.appleBalance) : '0.0000';

  return (
    <div className="bg-surface rounded-lg border border-border p-4">
      <h2 className="text-sm font-medium text-white mb-4">User Trading</h2>

      {/* Balance Display */}
      <div className="mb-4 space-y-2">
        <div className="bg-surface-light rounded-lg p-3">
          <div className="text-xs text-gray-400 mb-1">ETH Balance</div>
          <div className="text-lg font-semibold text-white">{ethBalance} ETH</div>
        </div>
        <div className="bg-surface-light rounded-lg p-3">
          <div className="text-xs text-gray-400 mb-1">APPL Balance</div>
          <div className="text-lg font-semibold text-white">{appleBalance} APPL</div>
        </div>
      </div>

      {/* Error Messages */}
      {error && (
        <div className="mb-4 p-3 bg-red-500/20 border border-red-500/50 rounded text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* Buy Section */}
      <div className="mb-4">
        <label className="block text-sm text-gray-400 mb-1">Buy APPL with ETH</label>
        <div className="flex space-x-2">
          <input
            type="number"
            value={buyAmount}
            onChange={(e) => setBuyAmount(e.target.value)}
            placeholder="0.0"
            disabled={isLoading}
            step="0.01"
            min="0"
            className="flex-1 bg-surface-light border border-border rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 disabled:opacity-50"
          />
          <button
            onClick={handleBuy}
            disabled={isLoading || !buyAmount}
            className="bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
          >
            Buy
          </button>
        </div>
      </div>

      {/* Sell Section */}
      <div>
        <label className="block text-sm text-gray-400 mb-1">Sell APPL for ETH</label>
        <div className="flex space-x-2">
          <input
            type="number"
            value={sellAmount}
            onChange={(e) => setSellAmount(e.target.value)}
            placeholder="0.0"
            disabled={isLoading}
            step="0.01"
            min="0"
            className="flex-1 bg-surface-light border border-border rounded px-3 py-2 text-white focus:outline-none focus:border-blue-500 disabled:opacity-50"
          />
          <button
            onClick={handleSell}
            disabled={isLoading || !sellAmount}
            className="bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-medium py-2 px-4 rounded transition"
          >
            Sell
          </button>
        </div>
      </div>
      <Toast
        visible={showToast}
        message={toastMessage}
        onClose={() => setShowToast(false)}
      />
    </div>
  );
}
