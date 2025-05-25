document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('signForm');
    const resultDiv = document.getElementById('result');
    const signatureInfo = document.getElementById('signatureInfo');
    const loader = document.createElement('div');
    loader.className = 'loader';

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        form.appendChild(loader);
        loader.style.display = 'block';
        resultDiv.classList.add('hidden');

        const dataFile = document.getElementById('dataFile').files[0];
        const keyFile = document.getElementById('keyFile').files[0];

        if (!dataFile || !keyFile) {
            alert('Пожалуйста, выберите файл и приватный ключ');
            loader.style.display = 'none';
            return;
        }

        try {
            const [dataContent, keyContent] = await Promise.all([
                readFileAsArrayBuffer(dataFile),
                readFileAsText(keyFile)
            ]);

            const formData = new FormData();
            formData.append('data', new Blob([dataContent]));
            formData.append('key', keyContent);

            const response = await fetch('/api/sign', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error(await response.text());
            }

            const { signature } = await response.json();

            const blob = new Blob([signature], { type: 'application/octet-stream' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${dataFile.name}.sig`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            signatureInfo.innerHTML = `
                <p>Файл успешно подписан!</p>
                <p>Размер подписи: ${signature.length} байт</p>
            `;
            resultDiv.classList.remove('hidden');

        } catch (err) {
            alert('Ошибка: ' + err.message);
            console.error(err);
        } finally {
            loader.style.display = 'none';
        }
    });

    function readFileAsArrayBuffer(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = () => resolve(reader.result);
            reader.onerror = reject;
            reader.readAsArrayBuffer(file);
        });
    }

    function readFileAsText(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = () => resolve(reader.result);
            reader.onerror = reject;
            reader.readAsText(file);
        });
    }
});