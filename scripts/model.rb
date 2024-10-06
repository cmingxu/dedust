
$fee = 0.0025

class Model
  attr_accessor :x, :y, :k, :tin, :limit, :tin_after_fee
  def initialize(x, y, tin, limit)
    @x = x
    @y = y
    @tin = tin
    @limit = limit

    @k = x * y
    @tin_after_fee = tin * (1 - $fee)
  end

  def actual_tout
    @y - (@k / (@tin_after_fee + @x))
  end

  def limit_actual_out_ratio
    @limit / actual_tout
  end

  def if_bot_buy_amount(amount)
    # bot buy
    bot_ton_in = amount * (1 - $fee)
    bot_jetton_out = @y - (@k / (@x + bot_ton_in))
    y_after_bot_buy = @y - bot_jetton_out
    x_after_bot_buy = @x + bot_ton_in

    # trade buy
    trade_jetton_out = y_after_bot_buy - (@k / (@tin_after_fee + x_after_bot_buy))
    limit_vs_trade_jetton_out = @limit / (trade_jetton_out * 1.0)
    x_after_trade_buy = x_after_bot_buy + @tin_after_fee
    y_after_trade_buy = y_after_bot_buy - trade_jetton_out

    # bot sell
    bot_sell_jetton_after_fee = bot_jetton_out * (1 - $fee)
    bot_ton_out  = x_after_trade_buy - (@k / (y_after_trade_buy + bot_sell_jetton_after_fee))

    return bot_jetton_out, trade_jetton_out, limit_vs_trade_jetton_out, bot_ton_out
  end
end

def main
  model = Model.new 7602922176642667, 40814970349566, 3000000000, 15713789
  puts "actual out: #{model.actual_tout}"
  puts "limit actual out ratio: #{model.limit_actual_out_ratio}"


  1.upto(10).each do |i|
    puts "===================="
    bot_ton_in = 1000000000 * i
    bot_actual_out, trade_actual_out, limit_vs_trade_actual_out, bot_ton_out = model.if_bot_buy_amount(bot_ton_in)
    puts "bot ton in: #{bot_ton_in}"
    puts "bot jetton out: #{bot_actual_out}"
    puts "trade jetton out: #{trade_actual_out}"
    puts "limit vs trade jetton out: #{limit_vs_trade_actual_out}"
    puts "bot ton out: #{bot_ton_out}"
    puts "bot profit: #{bot_ton_out - bot_ton_in}, #{ (bot_ton_out - bot_ton_in) /  10 ** 9}"
  end
end

main
