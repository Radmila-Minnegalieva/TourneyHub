import { useCallback, useEffect, useRef, useState } from "react";
import { errorMessage } from "../api/client";
export function useResource<T>(
  loader: () => Promise<T>,
  deps: readonly unknown[] = [],
) {
  const [data, setData] = useState<T>();
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0);
  const identity = useRef(deps);
  const backgroundRequested = useRef(false);
  const reload = useCallback(() => {
    backgroundRequested.current = false;
    setRevision((n) => n + 1);
  }, []);
  const refresh = useCallback(() => {
    backgroundRequested.current = true;
    setRevision((n) => n + 1);
  }, []);
  useEffect(() => {
    let current = true;
    const background =
      backgroundRequested.current &&
      deps.length === identity.current.length &&
      deps.every((value, index) => Object.is(value, identity.current[index]));
    identity.current = deps;
    backgroundRequested.current = false;
    setError("");
    if (!background) {
      setLoading(true);
      setData(undefined);
    }
    loader()
      .then((value) => {
        if (current) setData(value);
      })
      .catch((e) => {
        if (current) setError(errorMessage(e));
      })
      .finally(() => {
        if (current) setLoading(false);
      });
    return () => {
      current = false;
    };
    // The caller supplies the resource identity, like useEffect dependencies.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, revision]);
  return { data, error, loading, reload, refresh };
}
