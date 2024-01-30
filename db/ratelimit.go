package db

import (
	"database/sql"
	"eth2-exporter/utils"
)

func GetUserRatelimitProduct() {}

func UpdateApiRatelimits() (sql.Result, error) {
	return FrontendWriterDB.Exec(
		`with 
			stripe_price_ids as (
				select product, price_id from ( values
					('sapphire', $1),
					('emerald',  $2),
					('diamond',  $3),
					('custom1',  $4),
					('custom2',  $5),
					('whale',    $6),
					('goldfish', $7),
					('plankton', $8)
				) as x(product, price_id)
			),
			current_api_products as (
				select distinct on (product) product, second, hour, month, valid_from 
				from api_products 
				where valid_from <= now()
				order by product, valid_from desc
			)
		insert into api_ratelimits (user_id, second, hour, month, valid_until, changed_at)
		select
			u.id as user_id,
			greatest(coalesce(cap1.second,0),coalesce(cap2.second,0)) as second,
			greatest(coalesce(cap1.hour  ,0),coalesce(cap2.hour  ,0)) as hour,
			greatest(coalesce(cap1.month ,0),coalesce(cap2.month ,0)) as month,
			to_timestamp('3000-01-01', 'YYYY-MM-DD') as valid_until,
			now() as changed_at
		from users u
			left join users_stripe_subscriptions uss on uss.customer_id = u.stripe_customer_id and uss.active = true
			left join stripe_price_ids spi1 on spi1.price_id = uss.price_id
			left join current_api_products cap1 on cap1.product = coalesce(spi1.product,'free')
			left join app_subs_view asv on asv.user_id = u.id and asv.active = true
			left join current_api_products cap2 on cap2.product = coalesce(asv.product_id,'free')
		on conflict (user_id) do update set
			second = excluded.second,
			hour = excluded.hour,
			month = excluded.month,
			valid_until = excluded.valid_until,
			changed_at = now()
		where
			api_ratelimits.second != excluded.second 
			or api_ratelimits.hour != excluded.hour 
			or api_ratelimits.month != excluded.month`,
		utils.Config.Frontend.Stripe.Sapphire,
		utils.Config.Frontend.Stripe.Emerald,
		utils.Config.Frontend.Stripe.Diamond,
		utils.Config.Frontend.Stripe.Custom1,
		utils.Config.Frontend.Stripe.Custom2,
		utils.Config.Frontend.Stripe.Whale,
		utils.Config.Frontend.Stripe.Goldfish,
		utils.Config.Frontend.Stripe.Plankton,
	)
}
