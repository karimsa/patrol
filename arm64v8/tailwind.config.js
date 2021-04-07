module.exports = {
  purge: {
    mode: 'layers',
    enabled: process.env.NODE_ENV === 'production',
    preserveHtmlElements: false,
    content: ['./index.html'],
  },
}
