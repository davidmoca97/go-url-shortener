const input = document.getElementById('btn-submit');
const form = document.getElementById('form');
const btnClipboard = document.getElementById('copy-to-clipboard');
const newURL = document.getElementById('new-url');
const originalURL = document.getElementById('original-url');

const SHORTEN_END_POINT = 'url-shortener';

form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const url = input.value;
    if (!isURL(url)) {
        input.value = '';
        alert("Please provide a valid URL");
        return;
    }
    const body = { originalUrl: url }
    try {
        const res = await fetch(`/${SHORTEN_END_POINT}`, {
            method: "POST",
            body: JSON.stringify(body)
        });
        if (!res.ok && res.status < 500) {
            alert('Please provide a valid URL');
            return;
        }
        const data = await res.json();
        if (data !== undefined) {
            const { shortUrl, originalUrl } = data;
            newURL.innerHTML = shortUrl;
            originalURL.innerHTML = `<b>original URL</b>: ${originalUrl}`;
            input.value = '';
        }
    } catch (e) {
        alert("Server error, try again later");
    }
});

function isURL(str) {
    if (
        str.substr(0, 4).toUpperCase() !== "HTTP" &&
        str.substr(0, 5).toUpperCase() !== "HTTPS"
    ) {
        return false;
    }
    try {
        url = new URL(str);
    } catch (_) {
        return false;
    }
    return true;
}

input.addEventListener('click', () => input.select());
btnClipboard.addEventListener('click', () => {
    copyToClipboard(newURL.innerHTML);
});

function copyToClipboard(str) {
    const el = document.createElement('textarea');
    el.value = str;
    document.body.appendChild(el);
    el.select();
    document.execCommand('copy');
    document.body.removeChild(el);
}

