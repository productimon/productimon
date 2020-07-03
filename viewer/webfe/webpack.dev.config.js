const path = require('path');
console.log('path: ' + path.resolve(__dirname));

module.exports = {
  mode: 'development',
  entry: [path.resolve(__dirname, 'src/index.js')],
  resolve: { extensions: ['.tsx', '.ts', '.js'], symlinks: false },
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: { loader: 'babel-loader' },
      },
      { test: /\.tsx?$/, exclude: /node_modules/, use: 'ts-loader' },
      {
        test: /\.css$/,
        exclude: /node_modules/,
        use: ['style-loader', 'css-loader'],
      },
      { test: /\.(png|svg|jpg|gif)$/, use: ['file-loader'] },
    ],
  },
};
