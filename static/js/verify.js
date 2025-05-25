document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('verifyForm');
    const resultDiv = document.getElementById('result');
    const verificationResult = document.getElementById('verificationResult');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        // Очищаем предыдущие результаты
        resultDiv.classList.add('hidden');
        verificationResult.innerHTML = '';

        // Показываем индикатор загрузки
        const loadingIndicator = document.createElement('div');
        loadingIndicator.textContent = 'Проверяем подпись...';
        verificationResult.appendChild(loadingIndicator);

        const originalFile = document.getElementById('originalFile').files[0];
        const signatureFile = document.getElementById('signatureFile').files[0];
        const certFile = document.getElementById('certFile').files[0];

        if (!originalFile || !signatureFile || !certFile) {
            showError('Пожалуйста, выберите все необходимые файлы');
            return;
        }

        try {
            // Читаем файлы
            const [signatureData, certData] = await Promise.all([
                readFileAsText(signatureFile),
                readFileAsText(certFile)
            ]);

            // Создаем FormData
            const formData = new FormData();
            formData.append('data', originalFile);
            formData.append('signature', signatureData);
            formData.append('cert', certData);

            // Отправляем запрос
            const response = await fetch('/api/verify', {
                method: 'POST',
                body: formData
            });

            // Обрабатываем ответ
            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.error || 'Неизвестная ошибка');
            }

            // Показываем успешный результат
            showSuccess(result.message || 'Подпись действительна');

        } catch (err) {
            // Показываем ошибку
            showError(err.message);
            console.error('Ошибка проверки:', err);
        }
    });

    function readFileAsText(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = () => resolve(reader.result);
            reader.onerror = () => reject(new Error('Ошибка чтения файла'));
            reader.readAsText(file);
        });
    }

    function showSuccess(message) {
        resultDiv.classList.remove('hidden');
        verificationResult.innerHTML = `
            <div class="alert alert-success">
                <span class="success-icon">✓</span>
                ${message}
            </div>
        `;
    }

    function showError(message) {
        resultDiv.classList.remove('hidden');
        verificationResult.innerHTML = `
            <div class="alert alert-error">
                <span class="error-icon">✗</span>
                ${message}
            </div>
        `;
    }
});