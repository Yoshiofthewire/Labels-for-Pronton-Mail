import { useEffect, useState } from "react";
import { getJSON } from "../api/client";

type Health = {
  healthy: boolean;
  unhealthyForSeconds: number;
  failureReason: string[];
};

export function HealthPage() {
  const [health, setHealth] = useState<Health | null>(null);

  useEffect(() => {
    getJSON<Health>("/api/health").then(setHealth).catch(() => setHealth(null));
  }, []);

  return (
    <section className="panel">
      <h2>Health</h2>
      {health ? <pre>{JSON.stringify(health, null, 2)}</pre> : <p>Waiting for health data.</p>}
    </section>
  );
}
