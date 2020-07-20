import moment from "moment";
import React from "react";

import { grpc } from "@improbable-eng/grpc-web";
import { Empty } from "productimon/proto/common/common_pb";

// Number of miliseconds per unit
export const timeUnits = {
  Seconds: 10 ** 3,
  Minutes: 10 ** 3 * 60,
  Hours: 10 ** 3 * 60 * 60,
  Days: 10 ** 3 * 60 * 60 * 24,
  Weeks: 10 ** 3 * 60 * 60 * 24 * 7,
  Months: 10 ** 3 * 60 * 60 * 24 * 30,
  Years: 10 ** 3 * 60 * 60 * 24 * 365,
};

export function calculateDate(unit, val) {
  const mult = timeUnits[unit];
  return new Date().getTime() - val * mult;
}

export function parseDateTime(date, time) {
  const dateTime = date + "T" + time;
  return Date.parse(dateTime) * 10 ** 6;
}

export function nowNano() {
  return new Date() * 10 ** 6;
}

// set of colors taken from google charts
const google_colors = [
  "#3366cc",
  "#dc3912",
  "#ff9900",
  "#109618",
  "#990099",
  "#0099c6",
  "#dd4477",
  "#66aa00",
  "#b82e2e",
  "#316395",
  "#994499",
  "#22aa99",
  "#aaaa11",
  "#6633cc",
  "#e67300",
  "#8b0707",
  "#651067",
  "#329262",
  "#5574a6",
  "#3b3eac",
  "#b77322",
  "#16d620",
  "#b91383",
  "#f4359e",
  "#9c5935",
  "#a9c413",
  "#2a778d",
  "#668d1c",
  "#bea413",
  "#0c5922",
  "#743411",
];

const fixedColorMap = {
  Active: "#dc3912",
  Inactive: "#3366cc",
  Total: "#3366cc",
};

// colorMap is a universal mapping of label -> display color
const colorMap = new Map();
var colorIdx = 0;

export function getLabelColor(label) {
  if (fixedColorMap[label]) return fixedColorMap[label];
  if (!colorMap.has(label)) {
    colorMap.set(label, google_colors[colorIdx]);
    colorIdx++;
    colorIdx = colorIdx % google_colors.length;
  }
  return colorMap.get(label);
}

// format a time in seconds to readable string
export function humanizeDuration(seconds) {
  const duration = moment.duration(seconds * 10 ** 3);
  if (seconds < 60) {
    return `${duration.seconds()} seconds`;
  }
  return duration.humanize();
}

// TODO: make time format customisable
export function toSec(nanoseconds) {
  return nanoseconds / 10 ** 9;
}

export function redirectToLogin() {
  window.localStorage.removeItem("token");
  if (location.pathname != "/") location.href = "/";
}

export function rpc(methodDescriptor, request) {
  const token = window.localStorage.getItem("token");
  if (!request) request = new Empty();
  return new Promise((resolve, reject) => {
    grpc.unary(methodDescriptor, {
      host: "/rpc",
      metadata: new grpc.Metadata({ Authorization: token }),
      onEnd: ({ status, statusMessage, headers, message }) => {
        if (status != 0) {
          console.error("Error on RPC:", status, statusMessage);
          reject(statusMessage);
        }
        resolve(message);
      },
      request,
    });
  });
}

export function formatNano(ns, format = "lll") {
  return moment(toSec(ns), "X").format(format);
}
