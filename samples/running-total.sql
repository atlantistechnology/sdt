SELECT
DATE_FORMAT(co.order_date, '%Y-%m') AS order_month,
DATE_FORMAT(co.order_date, '%Y-%m-%d') AS order_day,
COUNT(DISTINCT co.order_id) AS num_orders,
COUNT(ol.book_id) AS num_books,
SUM(ol.price) AS total_price,
SUM(COUNT(ol.book_id)) OVER (
  PARTITION BY DATE_FORMAT(co.order_date, '%Y-%m')
  ORDER BY DATE_FORMAT(co.order_date, '%Y-%m-%d')
) AS running_total_num_books
FROM cust_order co
INNER JOIN order_line ol ON co.order_id = ol.order_id
GROUP BY 
  DATE_FORMAT(co.order_date, '%Y-%m'),
  DATE_FORMAT(co.order_date, '%Y-%m-%d')
ORDER BY co.order_date ASC;
