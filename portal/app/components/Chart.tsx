import {ChartResults} from "@/services/data/benchResults";
import {colors} from "@/lib/colors";
import dayjs from "dayjs";

// build時にwindow変数への参照を回避するため、apexchartsは動的ロードする
import dynamic from "next/dynamic";
const ApexChart = dynamic(() => import("react-apexcharts"), {ssr: false});

type ChartProps = {
  data: ChartResults
}

export default function Chart(props: ChartProps) {
  return (
    <ApexChart
      options={{
        chart: {
          animations: {
            enabled: false,
          },
        },
        colors,
        stroke: {
          width: 2,
        },
        markers: {
          size: 3,
          strokeWidth: 0.5,
        },
        xaxis: {
          type: "datetime",
          labels: {
            formatter: (value, timestamp) => {
              return dayjs(timestamp).format("MM/DD HH:mm:ss");
            },
          },
        },
        yaxis: {
          title: {
            text: "Score",
          },
          labels: {
            formatter: (val, opts) => {
              return val.toLocaleString();
            },
          },
        },
        legend: {
          onItemClick: {},
        },
      }}
      series={props.data}
      zoom={{
        type: "x",
        enabled: true,
      }}
      type="line"
      height="100%"
    />
  );
}
