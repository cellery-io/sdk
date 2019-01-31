import ballerina/io;
import celleryio/cellery;

cellery:Component helloWorldComp = {
    name: "HelloWorld",
    source: {
        image: "sumedhassk/hello-world:1.0.0"
    },
    ingresses: {
        "hello": new cellery:HTTPIngress(9090, "hello",
            [
                {
                    path: "/",
                    method: "GET"
                }
            ]
        )
    }
};

cellery:CellImage helloCell = new("Hello-World");

public function build() {
    helloCell.addComponent(helloWorldComp);

    helloCell.exposeGlobalAPI(helloWorldComp);

    var out = cellery:createImage(helloCell);
    if (out is boolean) {
        io:println("Hello Cell Built successfully.");
    }
}

