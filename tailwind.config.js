/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./templates/**/*.{html,js}",
    ],
    theme: {
        extend: {
            colors: {
                'background': '#2b2b2b',
                'primary':  '#48e596',
                'secondary': '#134e27',
                'accent': '#ffa348',
            },
            fontFamily: {
                sans: ['"Inter"', 'monospace']
            },
        },
    },
    plugins: [
        function ({ addVariant }) {
            addVariant('child', '& > *');
        }
    ],
}