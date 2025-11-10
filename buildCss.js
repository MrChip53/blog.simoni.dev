const fs = require('fs').promises;
const path = require('path');
const { spawnSync } = require('child_process');


console.log('Building theme CSS files...');

// buildThemes.js
// Scans templates/**/ for .css files and runs Tailwind for each to css/**/


const root = path.resolve(__dirname);
const themesDir = path.join(root, 'templates');
const outBase = path.join(root, 'css');
const tailwindConfig = '../tailwind.config.js'; // relative to themesDir

async function walk(dir) {
    const entries = await fs.readdir(dir, { withFileTypes: true });
    const files = await Promise.all(entries.map(async (e) => {
        const res = path.join(dir, e.name);
        return e.isDirectory() ? await walk(res) : res;
    }));
    return files.flat();
}

(async () => {
    try {
        const all = await walk(themesDir);
        const cssFiles = all.filter(f => f.endsWith('.css'));
        if (cssFiles.length === 0) {
            console.log('No .css files found in', themesDir);
            return;
        }

        for (const abs of cssFiles) {
            const rel = path.relative(themesDir, abs);
            const outPath = path.join(outBase, rel);
            await fs.mkdir(path.dirname(outPath), { recursive: true });

            console.log('Building', rel, '->', path.relative(root, outPath));

            // run tailwind via npx so it uses local install if present
            const args = ['tailwindcss', '-c', tailwindConfig, '-i', rel, '-o', path.relative(themesDir, outPath)];
            const res = spawnSync('npx', args, { cwd: themesDir,stdio: 'inherit', env: process.env, shell: true });

            if (res.error) throw res.error;
            if (res.status !== 0) throw new Error(`tailwind exit code ${res.status} for ${rel}`);
        }

        console.log('Done.');
    } catch (err) {
        console.error('Error:', err.message || err);
        process.exit(1);
    }
})();