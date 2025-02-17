import {Fragment} from "react";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import {RankingResults} from "@/services/data/benchResults";

type RankingProps = {
  data: RankingResults
}

export default function Ranking(props: RankingProps) {
  return (
    <Fragment>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Rank</TableCell>
            <TableCell>Team</TableCell>
            <TableCell>Score</TableCell>
            <TableCell>Lang</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {props.data.map(row => (
            <TableRow key={row.team}>
              <TableCell>{row.rank}</TableCell>
              <TableCell>{row.team}</TableCell>
              <TableCell>{row.score.toLocaleString()}</TableCell>
              <TableCell>{row.lang}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Fragment>
  );
}
