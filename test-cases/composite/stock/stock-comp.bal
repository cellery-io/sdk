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

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:Composite stockComposite = check cellery:constructImage(untaint iName);
    return cellery:createInstance(stockComposite, iName, instances, startDependencies, shareDependencies);
}
