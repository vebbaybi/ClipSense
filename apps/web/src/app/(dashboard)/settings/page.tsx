// SettingsPage shows static configuration hints for the MVP.
export default function SettingsPage() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Settings</h1>
      <div className="card space-y-2">
        <p className="text-slate-300">API URL</p>
        <p className="text-slate-400 text-sm">Configure via NEXT_PUBLIC_API_URL env when starting the app.</p>
      </div>
    </div>
  );
}
