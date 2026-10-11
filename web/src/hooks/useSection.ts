import { useEffect, useState } from "react";

export type Section = "overview" | "catalog" | "hr" | "data";
function section(): Section {
  const value = window.location.pathname.slice(1);
  return value === "catalog" || value === "hr" || value === "data"
    ? value
    : "overview";
}
export function useSection(): [Section, (next: Section) => void] {
  const [current, setCurrent] = useState(section);
  useEffect(() => {
    const sync = () => setCurrent(section());
    window.addEventListener("popstate", sync);
    return () => window.removeEventListener("popstate", sync);
  }, []);
  const navigate = (next: Section) => {
    if (next === section()) return;
    window.history.pushState(null, "", next === "overview" ? "/" : `/${next}`);
    setCurrent(next);
  };
  return [current, navigate];
}
