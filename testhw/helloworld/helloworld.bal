import celleryio/cellery;

cellery:Component helloWorldComp = {
    name: "hello-world",
    source: {
        image: "sumedhassk/hello-world:1.0.0"
    },
    ingresses: {
        hello: new cellery:HttpApiIngress(
                   9090,

                   "hello",
                   [
                       {
                           path: "/*",
                           method: "GET"
                       }
                   ]

        )
    }
};

cellery:CellImage helloCell = new();

public function build(string orgName, string imageName, string imageVersion) {
    helloCell.addComponent(helloWorldComp);

    helloCell.exposeGlobal(helloWorldComp);

    _= cellery:createImage(helloCell, orgName, imageName, imageVersion);
}

