/* tslint:disable */
/* eslint-disable */
//  This file was automatically generated and should not be edited.

export type BenchmarkResults = {
  __typename: "BenchmarkResults",
  id: string,
  team: string,
  bench_status?: string | null,
  lang?: string | null,
  queued_at?: string | null,
  started_at?: string | null,
  ended_at?: string | null,
  score?: number | null,
};

export type TableBenchmarkResultsFilterInput = {
  id?: TableStringFilterInput | null,
  team?: TableStringFilterInput | null,
  bench_status?: TableStringFilterInput | null,
  lang?: TableStringFilterInput | null,
  queued_at?: TableStringFilterInput | null,
  started_at?: TableStringFilterInput | null,
  ended_at?: TableStringFilterInput | null,
  score?: TableIntFilterInput | null,
};

export type TableStringFilterInput = {
  ne?: string | null,
  eq?: string | null,
  le?: string | null,
  lt?: string | null,
  ge?: string | null,
  gt?: string | null,
  contains?: string | null,
  notContains?: string | null,
  between?: Array< string | null > | null,
  beginsWith?: string | null,
  attributeExists?: boolean | null,
  size?: ModelSizeInput | null,
};

export type ModelSizeInput = {
  ne?: number | null,
  eq?: number | null,
  le?: number | null,
  lt?: number | null,
  ge?: number | null,
  gt?: number | null,
  between?: Array< number | null > | null,
};

export type TableIntFilterInput = {
  ne?: number | null,
  eq?: number | null,
  le?: number | null,
  lt?: number | null,
  ge?: number | null,
  gt?: number | null,
  between?: Array< number | null > | null,
  attributeExists?: boolean | null,
};

export type BenchmarkResultsConnection = {
  __typename: "BenchmarkResultsConnection",
  items?:  Array<BenchmarkResults | null > | null,
  nextToken?: string | null,
};

export type GetBenchmarkResultsQueryVariables = {
  id: string,
};

export type GetBenchmarkResultsQuery = {
  getBenchmarkResults?:  {
    __typename: "BenchmarkResults",
    id: string,
    team: string,
    bench_status?: string | null,
    lang?: string | null,
    queued_at?: string | null,
    started_at?: string | null,
    ended_at?: string | null,
    score?: number | null,
  } | null,
};

export type ListBenchmarkResultsQueryVariables = {
  filter?: TableBenchmarkResultsFilterInput | null,
  limit?: number | null,
  nextToken?: string | null,
};

export type ListBenchmarkResultsQuery = {
  listBenchmarkResults?:  {
    __typename: "BenchmarkResultsConnection",
    items?:  Array< {
      __typename: "BenchmarkResults",
      id: string,
      team: string,
      bench_status?: string | null,
      lang?: string | null,
      queued_at?: string | null,
      started_at?: string | null,
      ended_at?: string | null,
      score?: number | null,
    } | null > | null,
    nextToken?: string | null,
  } | null,
};
