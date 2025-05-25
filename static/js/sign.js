document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('signForm');
    const resultDiv = document.getElementById('result');
    const downloadBtn = document.getElementById('downloadSignature');
    const signatureInfo = document.getElementById('signatureInfo');

    let currentSignature = null;

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const dataFile = document.getElementById('dataFile').files[0];
        const keyFile = document.getElementById('keyFile').files[0];

        if (!dataFile || !keyFile) {
            alert('Пожалуйста, выберите оба файла');
            return;
        }

        try {
            // Читаем файлы как ArrayBuffer
            const [dataContent, keyContent] = await Promise.all([
                readFileAsArrayBuffer(dataFile),
                readFileAsText(keyFile)
            ]);

            // Отправляем на сервер
            const formData = new FormData();
            formData.append('data', new Blob([dataContent]));
            formData.append('key', keyContent);

            const response = await fetch('/api/sign', {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                const { signature } = await response.json();
                currentSignature = signature;

                // Показываем результат
                signatureInfo.innerHTML = `
                    <p>Файл успешно подписан!</p>
                    <p>Размер подписи: ${signature.length} байт</p>
                `;
                resultDiv.classList.remove('hidden');

                // Настраиваем кнопку скачивания
                downloadBtn.href = URL.createObjectURL(new Blob([signature], { type: 'application/octet-stream' }));
                downloadBtn.download = 'signature.sig';
            } else {
                const error = await response.json();
                throw new Error(error.details || error.error);
            }
        } catch (err) {
            alert(`Ошибка: ${err.message}`);
            console.error(err);
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