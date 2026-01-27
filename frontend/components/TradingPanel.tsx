'use client';

import { useState, useEffect } from 'react';
import { tradeBuy, tradeSell, getUserBalance } from '@/lib/api';
import type { UserBalance } from '@/types';

export function TradingPanel() {
  const [balance, setBalance] = useState<UserBalance | null>(null);
  const [buyAmount, setBuyAmount] = useState('');
  const [sellAmount, setSellAmount] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

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

  const formatWei = (wei: string): string => {
    const weiBigInt = BigInt(wei);
    const eth = Number(weiBigInt) / 1e18;
    return eth.toFixed(4);
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
      setSuccess(`Trade successful! Tx: ${result.txHash.slice(0, 10)}...`);
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
      setSuccess(`Trade successful! Tx: ${result.txHash.slice(0, 10)}...`);
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
      <h2 className="text-lg font-semibold text-white mb-4">User Trading</h2>

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

      {/* Error/Success Messages */}
      {error && (
        <div className="mb-4 p-3 bg-red-500/20 border border-red-500/50 rounded text-red-400 text-sm">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 p-3 bg-green-500/20 border border-green-500/50 rounded text-green-400 text-sm">
          {success}
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
    </div>
  );
}
