# Only to demonstrate enabling STOMP
FROM rabbitmq:3.13-management
RUN rabbitmq-plugins enable rabbitmq_stomp