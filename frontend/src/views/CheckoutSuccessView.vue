<!--
  CheckoutSuccessView - confirmation screen shown after a successful payment
  (route: /checkout-success). Static page; PaymentView redirects here on success.
-->
<template>
  <div class="success-page">
    <el-icon size="64" color="#3d8b3d"><CircleCheck /></el-icon>
    <h1>Заказ оформлен!</h1>
    <p>Мы обработаем ваш заказ в ближайшее время.</p>
    <div class="actions">
      <RouterLink to="/account/orders"><el-button type="primary">Мои заказы</el-button></RouterLink>
      <RouterLink to="/catalogue"><el-button>Продолжить покупки</el-button></RouterLink>
    </div>
  </div>
</template>

<style scoped>
.success-page { text-align: center; padding: 120px 24px; }
h1 { font-size: 32px; font-weight: 800; margin: 24px 0 12px; letter-spacing: 2px; }
p { color: var(--color-text-muted); margin-bottom: 40px; }
.actions { display: flex; gap: 16px; justify-content: center; }
</style>
