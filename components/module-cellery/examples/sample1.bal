import ballerina/io;
import celleryio/cellery;

public function main() {
    cellery:CellImage cell1 = {
        name:"MyCell",
        port:80,
        component:{
            name: "myComp",
            src:1,
            hello:"type"
        }

    };
    var v =cellery:createCellImage(cell1);
}
