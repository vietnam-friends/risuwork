/* tslint:disable */
/* eslint-disable */
// this is an auto generated file. This will be overwritten

import * as APITypes from "../API";
type GeneratedQuery<InputType, OutputType> = string & {
  __generatedQueryInput: InputType;
  __generatedQueryOutput: OutputType;
};

export const getBenchmarkResults = /* GraphQL */ `query GetBenchmarkResults($id: String!) {
  getBenchmarkResults(id: $id) {
    id
    team
    bench_status
    lang
    queued_at
    started_at
    ended_at
    score
    __typename
  }
}
` as GeneratedQuery<
  APITypes.GetBenchmarkResultsQueryVariables,
  APITypes.GetBenchmarkResultsQuery
>;
export const listBenchmarkResults = /* GraphQL */ `query ListBenchmarkResults(
  $filter: TableBenchmarkResultsFilterInput
  $limit: Int
  $nextToken: String
) {
  listBenchmarkResults(filter: $filter, limit: $limit, nextToken: $nextToken) {
    items {
      id
      team
      bench_status
      lang
      queued_at
      started_at
      ended_at
      score
      __typename
    }
    nextToken
    __typename
  }
}
` as GeneratedQuery<
  APITypes.ListBenchmarkResultsQueryVariables,
  APITypes.ListBenchmarkResultsQuery
>;
