class OrderService {
    constructor() {
        this.baseUrl = window.location.origin;
        this.init();
    }

    init() {
        this.bindEvents();
    }

    bindEvents() {
        const form = document.getElementById('search-form');
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.searchOrder();
        });
    }

    async searchOrder() {
        const orderId = document.getElementById('order-id').value.trim();
        const resultsSection = document.getElementById('results-section');
        const loadingElement = document.getElementById('loading');
        const errorElement = document.getElementById('error');
        const orderDetailsElement = document.getElementById('order-details');

        errorElement.style.display = 'none';
        orderDetailsElement.style.display = 'none';
        loadingElement.style.display = 'block';

        if (!orderId) {
            this.showError('Please enter an Order ID');
            loadingElement.style.display = 'none';
            return;
        }

        try {
            const response = await fetch(`${this.baseUrl}/api/order/${orderId}`);

            if (!response.ok) {
                if (response.status === 404) {
                    throw new Error('Order not found');
                }
                throw new Error('Failed to fetch order');
            }

            const order = await response.json();
            this.displayOrder(order);

        } catch (error) {
            this.showError(error.message);
        } finally {
            loadingElement.style.display = 'none';
        }
    }

    showError(message) {
        const errorElement = document.getElementById('error');
        const errorMessageElement = document.getElementById('error-message');

        errorMessageElement.textContent = message;
        errorElement.style.display = 'block';
    }

    displayOrder(order) {
        const orderDetailsElement = document.getElementById('order-details');
        orderDetailsElement.innerHTML = this.createOrderHTML(order);
        orderDetailsElement.style.display = 'block';
    }

    createOrderHTML(order) {
        return `
            <div class="order-section">
                <h3>Order Information</h3>
                <div class="detail-grid">
                    <div class="detail-item">
                        <div class="detail-label">Order UID</div>
                        <div class="detail-value">${order.order_uid}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Track Number</div>
                        <div class="detail-value">${order.track_number}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Entry</div>
                        <div class="detail-value">${order.entry}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Locale</div>
                        <div class="detail-value">${order.locale}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Customer ID</div>
                        <div class="detail-value">${order.customer_id}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Delivery Service</div>
                        <div class="detail-value">${order.delivery_service}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Date Created</div>
                        <div class="detail-value">${new Date(order.date_created).toLocaleString()}</div>
                    </div>
                </div>
            </div>

            <div class="order-section">
                <h3>Delivery Information</h3>
                <div class="detail-grid">
                    <div class="detail-item">
                        <div class="detail-label">Name</div>
                        <div class="detail-value">${order.delivery.name}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Phone</div>
                        <div class="detail-value">${order.delivery.phone}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Email</div>
                        <div class="detail-value">${order.delivery.email}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Address</div>
                        <div class="detail-value">${order.delivery.address}, ${order.delivery.city}, ${order.delivery.region} ${order.delivery.zip}</div>
                    </div>
                </div>
            </div>

            <div class="order-section">
                <h3>Payment Information</h3>
                <div class="detail-grid">
                    <div class="detail-item">
                        <div class="detail-label">Transaction</div>
                        <div class="detail-value">${order.payment.transaction}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Amount</div>
                        <div class="detail-value">$${(order.payment.amount / 100).toFixed(2)}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Currency</div>
                        <div class="detail-value">${order.payment.currency}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Provider</div>
                        <div class="detail-value">${order.payment.provider}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Bank</div>
                        <div class="detail-value">${order.payment.bank}</div>
                    </div>
                    <div class="detail-item">
                        <div class="detail-label">Payment Date</div>
                        <div class="detail-value">${new Date(order.payment.payment_dt * 1000).toLocaleString()}</div>
                    </div>
                </div>
            </div>

            <div class="order-section">
                <h3>Items (${order.items.length})</h3>
                <div class="items-grid">
                    ${order.items.map(item => `
                        <div class="item-card">
                            <div class="detail-grid">
                                <div class="detail-item">
                                    <div class="detail-label">Name</div>
                                    <div class="detail-value">${item.name}</div>
                                </div>
                                <div class="detail-item">
                                    <div class="detail-label">Brand</div>
                                    <div class="detail-value">${item.brand}</div>
                                </div>
                                <div class="detail-item">
                                    <div class="detail-label">Price</div>
                                    <div class="detail-value">$${(item.price / 100).toFixed(2)}</div>
                                </div>
                                <div class="detail-item">
                                    <div class="detail-label">Sale</div>
                                    <div class="detail-value">${item.sale}%</div>
                                </div>
                                <div class="detail-item">
                                    <div class="detail-label">Total Price</div>
                                    <div class="detail-value">$${(item.total_price / 100).toFixed(2)}</div>
                                </div>
                                <div class="detail-item">
                                    <div class="detail-label">Status</div>
                                    <div class="detail-value">${item.status}</div>
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new OrderService();
});