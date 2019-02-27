//Import modules for reading fil
var fs = require('fs');
const express = require('express');
const orderListFile = 'orderlist.json';
const app = express();
const port = 80;

let obj = null;
fs.readFile(orderListFile, 'utf8', function (err, data) {
    if(!err) {
        obj = JSON.parse(data);
    }
    else {
        obj = JSON.parse("{error detected:error}");
    }

});

app.get('/getOrders', (req, res) => res.send(obj));

app.listen(port, () => console.log(`Order-service app listening on port ${port}!`));


