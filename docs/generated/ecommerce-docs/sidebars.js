module.exports = {
    docs: [
        {
            type: 'doc',
            id: 'index',
            label: 'Home',
        },
        {
            type: 'category',
            label: 'Nodes',
            items: [
                'nodes/customer',
                'nodes/admin',
                'nodes/ecommerce-system',
                'nodes/api-gateway',
                'nodes/order-service',
                'nodes/inventory-service',
                'nodes/payment-service',
                'nodes/order-db',
                'nodes/inventory-db'
            ],
        },
        {
            type: 'category',
            label: 'Relationships',
            items: [
                'relationships/customer-interacts-gateway',
                'relationships/admin-interacts-gateway',
                'relationships/gateway-connects-order',
                'relationships/gateway-connects-inventory',
                'relationships/order-connects-db',
                'relationships/order-connects-payment',
                'relationships/order-connects-inventory',
                'relationships/inventory-connects-db',
                'relationships/ecommerce-system-composition'
            ],
        },
        {
            type: 'category',
            label: 'Flows',
            items: [
                'flows/order-processing-flow',
                'flows/inventory-check-flow'
            ],
        }
    ]
};
