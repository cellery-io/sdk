# Cellery Product Review Sample

The application shows information about user reviews for the set of products. Each of the review item shows product information, product category information, customer information, rating and review description.

This samples demonstrate following features of the cellery.

* Intra-Cell communication
* Inter-Cell communication
* Different Protocol usage

The Product Review Sample application has five microservices which are group into three seperate cells.


The component interaction diagram of the application is shown below.

![samples-productreview](samples-productreview.png)

## Invoking the API

    # To retrive data from memory  
    curl <reviews-gateway>/reviews-1/reviews -v -H "Authorization: Bearer $jwt"
    # To retrive data from database
    curl <reviews-gateway>/reviews-1/reviews?db=true -v -H "Authorization: Bearer $jwt"
