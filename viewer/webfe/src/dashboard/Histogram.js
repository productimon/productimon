import React from 'react';
import { useTheme } from '@material-ui/core/styles';
import {BarChart, Bar, CartesianGrid, XAxis, YAxis, Label, ResponsiveContainer, Tooltip } from 'recharts';
import Title from './Title';

// Generate Sales Data
function createData(time, amount) {
  return { time, amount };
}

const data = [
  createData('20/6', 5),
  createData('21/6', 2),
  createData('22/6', 8),
  createData('23/6', 3),
  createData('24/6', 10),
  createData('25/6', 11),
  createData('26/6', 4),
  createData('27/6', 7),
  createData('28/6', 5),
];

export default function Histogram() {
  const theme = useTheme();

  return (
    <React.Fragment>
      <Title>Computer Usage per day</Title>
      <ResponsiveContainer>
        <BarChart
          data={data}
          margin={{
            top: 16,
            right: 16,
            bottom: 0,
            left: 24,
          }}
        >
          <CartesianGrid strokeDasharray="3 3"/>
          <XAxis dataKey="time" stroke={theme.palette.text.secondary} />
          <YAxis stroke={theme.palette.text.secondary}>
            <Label
              angle={270}
              position="left"
              style={{ textAnchor: 'middle', fill: theme.palette.text.primary }}
            >
              Hours
            </Label>
          </YAxis>
          <Bar dataKey="amount" fill="#c8e6c9" stroke='#484848'/>
        </BarChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
