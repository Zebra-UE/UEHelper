const path = require('path');

module.exports = {
  mode: 'development', // 可以根据需要切换 'production' 或 'development'
  entry: './src/index.ts', // 项目入口文件
  output: {
    filename: 'bundle.js', // 打包后的文件名
    path: path.resolve(__dirname, 'dist'), // 输出目录为项目根目录的 'dist'
  },
  resolve: {
    // 告诉 Webpack 如何解析文件扩展名，使导入时可以省略 .ts 和 .js
    extensions: ['.ts', '.js'],
  },
  module: {
    rules: [
      {
        test: /\.ts$/, // 匹配所有 .ts 文件
        use: 'ts-loader', // 使用 ts-loader 进行处理
        exclude: /node_modules/, // 排除 node_modules 目录
      },
    ],
  },
  // 如果你的工具用于浏览器环境，这是必要的
  target: 'web', 
};