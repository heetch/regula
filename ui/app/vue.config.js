module.exports = {
  // serve static files from a relative url
  baseUrl: './',
  // proxy any unknown requests (requests that did not match a static file)
  devServer: {
    proxy: 'http://localhost:5331/ui',
  },
};
