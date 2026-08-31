// Toast.tsx - Ephemeral bottom-right notice (user trade confirmations).
//
// Multi-line messages (fill + ETH/APPL price) from TradingPanel. Auto-dismisses
// ~2.5s after it appears. The dismiss timer keys off visible + message only, so
// frequent parent re-renders (price ticks, balance polling during an active
// sim) don't keep resetting it and leave the toast stuck on screen.
//
// CONNECTIONS:
//  - Producer: components/TradingPanel.tsx buildToastMessage
//  - Trade API: lib/api.ts tradeBuy / tradeSell → handlers.go

import { useEffect, useRef } from 'react';

const DISMISS_MS = 2500;

interface ToastProps {
  visible: boolean;
  message: string;
  onClose: () => void;
}

export const Toast = ({ visible, message, onClose }: ToastProps) => {
  // Keep the latest onClose without making it a timer dependency.
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    if (!visible) return;
    const timer = setTimeout(() => onCloseRef.current(), DISMISS_MS);
    return () => clearTimeout(timer);
  }, [visible, message]);

  if (!visible) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <div className="bg-surface-light border border-border shadow-lg rounded-lg px-4 py-3 text-sm text-white max-w-md whitespace-pre-line font-mono leading-relaxed">
        {message}
      </div>
    </div>
  );
};
