//Import modules for reading fil
var fs = require('fs');
const express = require('express');
const customersFile = 'customerlist.json';
const app = express();
const port = 3000;

let obj = null;
fs.readFile(customersFile, 'utf8', function (err, data) {
    if(!err) {
        obj = JSON.parse(data);
    }
    else {
        obj = JSON.parse("{error detected:error}");
    }
});

app.get('/getCustomers', (req, res) => res.send(obj));

app.listen(port, () => console.log(`Customer-service app listening on port ${port}!`));


