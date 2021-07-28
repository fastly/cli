const path = require("path");
const webpack = require("webpack");

module.exports = {
  entry: "./src/index.js",
  optimization: {
    minimize: true
  },
  target: "webworker",
  output: {
    filename: "index.js",
    path: path.resolve(__dirname, "bin"),
    libraryTarget: "this",
  },
  plugins: [
    new webpack.ProvidePlugin({
      URL: "core-js/web/url",
    }),
  ],
};
