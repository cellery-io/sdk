import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Stock Component
    cellery:Component stockComponent = {
        name: "stock",
        source: {
            image: "wso2cellery/sampleapp-stock:0.3.0"
        },
        ingresses: {
            http:<cellery:HttpsPortIngress>{port: 8080}
        }
    };

    cellery:Composite stockComposite = {
        components: {
            stockComp: stockComponent
        }
    };
    return cellery:createImage(stockComposite, untaint iName);
}


