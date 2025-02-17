import Typography from "@mui/material/Typography";

export default function Footer(props: any) {
  return (
    <Typography variant="body2" color="text.secondary" align="center" {...props}>
      {'risuwork '}
      {new Date().getFullYear()}
      {'.'}
    </Typography>
  );
}