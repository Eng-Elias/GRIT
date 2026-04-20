import { AlertTriangle, Shield } from 'lucide-react';

interface BusFactorGaugeProps {
  busFactor: number;
}

export default function BusFactorGauge({ busFactor }: BusFactorGaugeProps) {
  const isRisky = busFactor <= 2;

  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-4 text-center">
      <h3 className="text-sm font-medium text-zinc-400 mb-3">Bus Factor</h3>
      <div className={`inline-flex items-center justify-center h-20 w-20 rounded-full border-4 ${
        isRisky ? 'border-red-500/60 bg-red-950/30' : 'border-green-500/60 bg-green-950/30'
      }`}>
        <span className={`text-3xl font-bold ${isRisky ? 'text-red-400' : 'text-green-400'}`}>
          {busFactor}
        </span>
      </div>
      <div className="mt-3 flex items-center justify-center gap-1.5 text-sm">
        {isRisky ? (
          <>
            <AlertTriangle className="h-4 w-4 text-red-400" />
            <span className="text-red-400">High risk — knowledge concentrated</span>
          </>
        ) : (
          <>
            <Shield className="h-4 w-4 text-green-400" />
            <span className="text-green-400">Healthy knowledge distribution</span>
          </>
        )}
      </div>
    </div>
  );
}
