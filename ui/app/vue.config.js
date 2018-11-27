module.exports = {
  // serve static files from relative url
  baseUrl: './',
  // proxy any unknown requests (requests that did not match a static file)
  devServer: {
    proxy: 'http://localhost:5331/ui',
  },
};
