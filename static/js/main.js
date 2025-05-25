// Обработчик генерации
document.getElementById('generate')?.addEventListener('click', async function() {
    try {
        const response = await fetch('/api/generate', {
            method: 'POST'
        });

        if (response.ok) {
            const downloadCert = document.getElementById('downloadCert');
            const downloadKey = document.getElementById('downloadKey');

            downloadCert.removeAttribute('disabled');
            downloadKey.removeAttribute('disabled');

            alert('Сертификат и ключ успешно сгенерированы!');
        } else {
            const error = await response.json();
            alert(`Ошибка: ${error.error}`);
        }
    } catch (err) {
        alert(`Ошибка сети: ${err.message}`);
    }
});