export function Feedback({
  loading,
  error,
  retry,
}: {
  loading?: boolean;
  error?: string;
  retry?: () => void;
}) {
  if (loading)
    return (
      <p role="status" className="py-6 text-slate-500">
        Загрузка…
      </p>
    );
  if (error)
    return (
      <div role="alert" className="my-4 rounded-xl bg-red-50 p-4 text-red-800">
        {error}
        {retry && (
          <button className="ml-4 underline" onClick={retry}>
            Повторить
          </button>
        )}
      </div>
    );
  return null;
}
