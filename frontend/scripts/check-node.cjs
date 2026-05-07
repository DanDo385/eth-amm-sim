const major = Number(process.versions.node.split('.')[0]);

if (Number.isNaN(major)) {
  console.error('Unable to parse Node version:', process.versions.node);
  process.exit(1);
}

if (major < 20 || major >= 25) {
  console.error('\n[frontend] Unsupported Node.js version:', process.versions.node);
  console.error('[frontend] Use Node 20.x or 22.x LTS (see frontend/.nvmrc).');
  console.error('[frontend] Example: nvm use 20 && npm install\n');
  process.exit(1);
}
