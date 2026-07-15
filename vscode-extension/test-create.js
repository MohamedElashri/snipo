const axios = require('axios');

const url = 'http://localhost:8080';
const token = 'f1648ad3fbce99521d2d27c701d4d3668842c132466ad57fd84abea894a9f493';

const client = axios.create({
    baseURL: url,
    headers: { 'Authorization': `Bearer ${token}` }
});

async function run() {
    try {
        const response = await client.post('/api/v1/snippets', {
            title: 'Test Snippet',
            description: 'Test Description',
            content: 'console.log("hello world");',
            language: 'javascript'
        });
        console.log('Success:', response.data);
    } catch (error) {
        console.log('Error Status:', error.response?.status);
        console.log('Error Data:', error.response?.data);
        console.log('Error Message:', error.message);
    }
}

run();
