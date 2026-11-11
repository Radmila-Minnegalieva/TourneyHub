import { allPages, api, type Schema } from "../api/client";
import { teamDetails } from "../api/catalog";
import { useResource } from "./useResource";
export function useTeams() {
  const result = useResource(() => allPages<Schema["Team"]>("/users/me/teams"));
  return {
    ...result,
    teams: result.data || [],
    createTeam: async (name: string) => {
      await api("/teams", { method: "POST", body: { name } });
      result.reload();
    },
  };
}
export function useTeam(id: string) {
  return useResource(() => teamDetails(id), [id]);
}
