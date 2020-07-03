import React from 'react';
import { useTheme } from '@material-ui/core/styles';
import {
  PieChart, Pie, Sector, Cell, ResponsiveContainer, Tooltip
} from 'recharts';
import Title from './Title';

function createData(name, value) {
  return {name, value};
}

const data = [
  createData('Messenger', 4),
  createData('Reddit', 3),
  createData('Spotify', 3),
  createData('Podcasts', 2),
];


export default function DisplayPie() {
  const theme = useTheme();

  return (
    <React.Fragment>
      <Title>Time Spent on Andriod</Title>
      <ResponsiveContainer>
        <PieChart width={200} height={200}>
          {/*<Pie isAnimationActive={false} data={data} cx={200} cy={200} outerRadius={80} fill="#8884d8" label/>*/}
          <Pie data={data} cx={200} cy={100} innerRadius={40} outerRadius={80} fill="#c8e6c9" stroke='#484848'/>
          <Tooltip/>
        </PieChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
