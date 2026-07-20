
import { useEffect } from 'react';

interface ToastProps {
  visible: boolean;
  message: string;
  onClose: () => void;
}

export const Toast = ({ visible, message, onClose }: ToastProps) => {
  useEffect(() => {
    if (!visible) return;
    const timer = setTimeout(onClose, 5000);
    return () => clearTimeout(timer);
  }, [visible, onClose]);

  if (!visible) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <div className="bg-surface-light border border-border shadow-lg rounded-lg px-4 py-3 text-sm text-white max-w-md whitespace-pre-line font-mono leading-relaxed">
        {message}
      </div>
    </div>
  );
};
