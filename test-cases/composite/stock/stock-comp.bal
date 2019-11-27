import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    //Stock Component
    cellery:Component stockComponent = {
        name: "stock",
        src: {
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
    return <@untainted> cellery:createImage(stockComposite, iName);
}

public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage|cellery:Composite stockComposite = cellery:constructImage(iName);
    return <@untainted> cellery:createInstance(stockComposite, iName, instances, startDependencies, shareDependencies);
}
