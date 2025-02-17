import { client } from "@/services/amplify";
import { listBenchmarkResults } from "@/services/amplify/graphql/queries";
import { BenchmarkResults } from "@/services/amplify/API";
import teams from "@/lib/teams.json";

export type ChartResults = {
  name: string,
  data: {
    x: string,
    y: number,
  }[]
}[]

export type RankingResults = {
  rank: number,
  team: string,
  score: number,
  lang: string,
}[]

export type DataSet = {
  chartResults: ChartResults,
  rankingResults: RankingResults
}

export async function fetch(): Promise<DataSet> {
  const res = await client.graphql({
    query: listBenchmarkResults,
    variables: {
      filter: {
        ended_at: {"attributeExists": true},
        // 公開用ポータルでは運営チームのスコアは非表示
        team: {"notContains": "unnei"},
      },
      limit: 10000,
    },
  });
  const items = (res.data.listBenchmarkResults?.items || []) as BenchmarkResults[];
  items.forEach(item => item.team = (teams as { [key: string]: string })[item.team] || item.team);
  const groupedItems = Object.groupBy(items, item => item.team) as Record<string, BenchmarkResults[]>;

  return {
    chartResults: transformChartResults(groupedItems),
    rankingResults: transformRankingResults(groupedItems),
  };
}

function transformChartResults(groupedItems: Record<string, BenchmarkResults[]>): ChartResults {
  return Object.entries(groupedItems).map(([team, items]) => {
    return {
      name: team,
      data: items.sort((a, b) => {
        return (a.ended_at ?? "").localeCompare(b.ended_at ?? "");
      })
        .map(item => {
          return {
            y: item.score || 0,
            x: item.ended_at || "",
          };
        }) || [],
    };
  });
}

function transformRankingResults(groupedItems: Record<string, BenchmarkResults[]>): RankingResults {
  return Object.entries(groupedItems)
    .map(([team, items]) => {
      const latestItem = items[items.length - 1];
      // 最新スコアから+10%以内での最大値がランキングスコアとなる
      const rankingScore = Math.max(...items.filter(item => (item.score || 0) <= (latestItem.score || 0) * 1.1).map(item => item.score || 0));
      return {
        team,
        score: rankingScore,
        lang: latestItem?.lang || "",
      };
    })
    .sort((a, b) => {
      return (b.score ?? 0) - (a.score ?? 0);
    })
    .map((item, i) => {
      return {
        rank: i + 1,
        team: item.team,
        score: item.score,
        lang: item.lang,
      };
    });
}
