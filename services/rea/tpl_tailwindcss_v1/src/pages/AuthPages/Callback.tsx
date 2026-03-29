import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { handleCallback } from "../../auth/authProvider";
import LoadingSpinner from "../../components/ui/loading/LoadingSpinner";

export default function Callback() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    handleCallback()
      .then(() => {
        navigate("/projects", { replace: true });
      })
      .catch((err: unknown) => {
        const message =
          err instanceof Error ? err.message : "Authentication failed";
        setError(message);
        // Redirect to login after a short delay so the user can read the error
        setTimeout(() => {
          navigate("/login?error=callback_failed", { replace: true });
        }, 3000);
      });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen bg-white dark:bg-gray-900">
        <div className="text-center space-y-4 max-w-sm px-6">
          <p className="text-red-500 dark:text-red-400 text-sm">{error}</p>
          <p className="text-gray-500 dark:text-gray-400 text-xs">
            Redirecting to login...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center h-screen bg-white dark:bg-gray-900">
      <div className="text-center space-y-4">
        <LoadingSpinner text="Completing sign in..." />
      </div>
    </div>
  );
}
